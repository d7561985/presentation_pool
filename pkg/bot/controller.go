package bot

import (
	"fmt"
	"github.com/pkg/errors"
	"presentation_pool/pkg/models"
)

func (b *Bot) StartStep(vote *models.Vote) error {
	if int(b.status.Step) >= len(vote.Steps) {
		return fmt.Errorf("vote %q contain wrong num: %d", vote.Name, len(vote.Steps))
	}

	b.status.VoteName = vote.Name
	b.status.Status = models.StatusInProgress
	b.vote = vote
	b.status.Step = 0

	if err := b.store.SaveStatus(b.status); err != nil {
		return errors.WithMessage(err, "cant save status")
	}

	msg, err := b.msgShowCurrentStepWindow(0)
	if err != nil {
		return errors.WithMessage(err, "prepare step window")
	}

	return b.Broadcast(msg)
}

func (b *Bot) CompleteStep() error {
	b.status.Status = models.StatusComplete

	if err := b.store.SaveStatus(b.status); err != nil {
		return errors.WithMessage(err, "cant save status")
	}

	return nil
}

func (b *Bot) NextStep() error {
	if b.status.Status != models.StatusComplete {
		return fmt.Errorf("previouse step not completed")
	}

	b.status.Step++
	b.status.Status = models.StatusInProgress

	if err := b.store.SaveStatus(b.status); err != nil {
		return errors.WithMessage(err, "cant save status")
	}

	msg, err := b.msgShowCurrentStepWindow(0)
	if err != nil {
		return errors.WithMessage(err, "prepare step window")
	}

	return b.Broadcast(msg)
}
