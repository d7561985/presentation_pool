package excel

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"presentation_pool/pkg/models"
	"strconv"
	"strings"
)

type Excel struct {
	spreadsheetId string

	service *spreadsheet.Service
}

func New(ctx context.Context, sheetID string, creds []byte) (*Excel, error) {
	conf, err := google.JWTConfigFromJSON(creds, spreadsheet.Scope)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to retrieve Sheets client")
	}

	client := conf.Client(ctx)
	service := spreadsheet.NewServiceWithClient(client)

	return &Excel{
		spreadsheetId: sheetID,
		service:       service,
	}, nil
}

func (e *Excel) getSheet(title string) (*spreadsheet.Sheet, error) {
	sheet, err := e.service.FetchSpreadsheet(e.spreadsheetId)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to retrieve Sheets client")
	}

	sh, err := sheet.SheetByTitle(title)
	if err != nil {
		// can afford this
		if err = e.service.AddSheet(&sheet, spreadsheet.SheetProperties{
			Title: title,
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		return e.getSheet(title)
	}

	return sh, nil
}

var (
	ErrNotFound = errors.New("not found")
)

func (e *Excel) GetUser(id string) (*models.User, error) {
	sheet, err := e.getSheet("users")
	if err != nil {
		return nil, errors.WithMessagef(err, "cant get sheet")
	}

	for _, rows := range sheet.Rows {
		if rows == nil {
			continue
		}

		if rows[0].Value != id {
			continue
		}

		return toUser(rows), nil
	}

	return nil, ErrNotFound
}

func (e *Excel) SaveUser(u *models.User) error {
	sheet, err := e.getSheet("users")
	if err != nil {
		return errors.WithMessagef(err, "cant get sheet")
	}

	id := len(sheet.Rows)

	sheet.Update(id, 0, fmt.Sprint(u.ID))
	sheet.Update(id, 1, u.Email)
	sheet.Update(id, 2, u.UserName)
	sheet.Update(id, 3, u.FirstName)
	sheet.Update(id, 4, u.LastName)
	sheet.Update(id, 5, u.LanguageCode)
	sheet.Update(id, 6, u.IsBot)

	return errors.WithStack(sheet.Synchronize())
}

const (
	settingSuffix = "settings_"
)

func (e *Excel) GetUsers() ([]*models.User, error) {
	sheet, err := e.getSheet("users")
	if err != nil {
		return nil, errors.WithMessagef(err, "cant get sheet")
	}

	var res []*models.User
	for _, row := range sheet.Rows {
		res = append(res, toUser(row))
	}

	return res, nil
}

func (e *Excel) GetAllVotes() ([]models.Vote, error) {
	sheet, err := e.service.FetchSpreadsheet(e.spreadsheetId)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to retrieve Sheets client")
	}

	var res []models.Vote

	for _, sh := range sheet.Sheets {
		if !strings.HasPrefix(sh.Properties.Title, settingSuffix) {
			continue
		}

		vote := models.Vote{
			Name:  strings.Trim(sh.Properties.Title, settingSuffix),
			Steps: nil,
		}

		if len(sh.Rows) < 2 {
			continue
		}

		for _, rows := range sh.Rows[1:] {
			if len(rows) < 2 {
				continue
			}

			step := models.Step{
				Question: rows[0].Value,
				Option:   nil,
			}

			for _, row := range rows[1:] {
				step.Option = append(step.Option, row.Value)
			}

			vote.Steps = append(vote.Steps, step)
		}

		res = append(res, vote)
	}

	return res, nil
}

func (e *Excel) SaveStatus(in *models.StatusData) error {
	sheet, err := e.getSheet("current")
	if err != nil {
		return errors.WithMessagef(err, "cant get sheet")
	}

	sheet.Update(1, 0, in.VoteName)
	sheet.Update(1, 1, fmt.Sprintf("%d", in.Step))
	sheet.Update(1, 2, in.Status)

	return errors.WithStack(sheet.Synchronize())
}

func (e *Excel) GetStatus() (*models.StatusData, error) {
	sheet, err := e.getSheet("current")
	if err != nil {
		return nil, errors.WithMessagef(err, "cant get sheet")
	}

	res := &models.StatusData{Status: models.StatusIdle}
	if len(sheet.Rows) < 2 {
		return res, nil
	}

	res.VoteName = sheet.Rows[1][0].Value
	step, err := strconv.ParseInt(sheet.Rows[1][1].Value, 10, 64)
	if err != nil {
		return nil, errors.WithMessagef(err, "cant parse step")
	}

	res.Step = step
	res.Status = sheet.Rows[1][2].Value

	return res, nil
}

func toUser(rows []spreadsheet.Cell) *models.User {
	if len(rows) < 7 {
		return &models.User{ID: "-1"}
	}

	var isAdmin bool
	if len(rows) > 7 {
		isAdmin, _ = strconv.ParseBool(rows[7].Value)
	}

	user := &models.User{
		ID:           rows[0].Value,
		Email:        rows[1].Value,
		UserName:     rows[2].Value,
		FirstName:    rows[3].Value,
		LastName:     rows[4].Value,
		LanguageCode: rows[5].Value,
		IsBot:        rows[6].Value,
		IsAdmin:      isAdmin,
	}

	return user
}

func (e *Excel) SaveUserVote(voteName string, step int64, question string, data string, user *models.User) error {
	sheet, err := e.getSheet("user_votes_" + strings.ToLower(voteName))
	if err != nil {
		return errors.WithMessagef(err, "cant get sheet")
	}

	sheet.Update(0, int(step+1), question)

	for id, row := range sheet.Rows {
		if row[0].Value == user.Email {
			e.saveVite(sheet, id, user, step, data)
			return sheet.Synchronize()
		}
	}

	id := len(sheet.Rows)

	e.saveVite(sheet, id, user, step, data)
	return sheet.Synchronize()
}

func (e *Excel) saveVite(sheet *spreadsheet.Sheet, id int, user *models.User, step int64, data string) {
	sheet.Update(id, 0, user.Email)
	sheet.Update(id, int(step+1), data)
}
