package db_repo

import (
	"context"
	"fmt"

	"github.com/applike/gosoline/pkg/mon"
	"github.com/hashicorp/go-multierror"
)

type CrudRepository interface {
	Create(ctx context.Context, value ModelBased) error
	Read(ctx context.Context, id *uint, out ModelBased) error
	Update(ctx context.Context, value ModelBased) error
	Delete(ctx context.Context, value ModelBased) error
}

type notifyingRepository struct {
	CrudRepository

	logger    mon.Logger
	notifiers NotificationMap
}

func NewNotifyingRepository(logger mon.Logger, base CrudRepository) *notifyingRepository {
	return &notifyingRepository{
		CrudRepository: base,
		logger:         logger,
		notifiers:      make(NotificationMap),
	}
}

func (r *notifyingRepository) AddNotifierAll(c Notifier) {
	for _, t := range NotificationTypes {
		r.AddNotifier(t, c)
	}
}

func (r *notifyingRepository) AddNotifier(t string, c Notifier) {
	if _, ok := r.notifiers[t]; !ok {
		r.notifiers[t] = make([]Notifier, 0)
	}

	r.notifiers[t] = append(r.notifiers[t], c)
}

func (r *notifyingRepository) Create(ctx context.Context, value ModelBased) error {
	if err := r.CrudRepository.Create(ctx, value); err != nil {
		return err
	}

	return r.doCallback(ctx, Create, value)
}

func (r *notifyingRepository) Update(ctx context.Context, value ModelBased) error {
	if err := r.CrudRepository.Update(ctx, value); err != nil {
		return err
	}

	return r.doCallback(ctx, Update, value)
}

func (r *notifyingRepository) Delete(ctx context.Context, value ModelBased) error {
	if err := r.CrudRepository.Delete(ctx, value); err != nil {
		return err
	}

	return r.doCallback(ctx, Delete, value)
}

func (r *notifyingRepository) doCallback(ctx context.Context, callbackType string, value ModelBased) error {
	if _, ok := r.notifiers[callbackType]; !ok {
		return nil
	}

	logger := r.logger.WithContext(ctx)
	var errors error

	for _, c := range r.notifiers[callbackType] {
		err := c.Send(ctx, callbackType, value)
		if err != nil {
			errors = multierror.Append(errors, err)
			logger.Warn("%T notifier errored out with: %v", c, err)
		}
	}

	if errors != nil {
		return fmt.Errorf("there were errors during execution of the callbacks for %s: %w", callbackType, errors)
	}

	return nil
}
