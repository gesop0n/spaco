package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// PreviewReviewScheduleは、結果ごとの次回予定日を、記録せずに返すユースケースである。
type PreviewReviewSchedule struct {
	repository IPreviewReviewScheduleRepository
	timeZones  ITimeZoneReader
	clock      Clock
	scheduler  scheduler.Scheduler
}

func NewPreviewReviewSchedule(
	repository IPreviewReviewScheduleRepository,
	timeZones ITimeZoneReader,
	clock Clock,
	reviewScheduler scheduler.Scheduler,
) (*PreviewReviewSchedule, error) {
	if err := requireDependencies(
		"preview review schedule",
		dependency{name: "repository", value: repository},
		dependency{name: "time zone reader", value: timeZones},
		dependency{name: "clock", value: clock},
	); err != nil {
		return nil, err
	}
	return &PreviewReviewSchedule{
		repository: repository,
		timeZones:  timeZones,
		clock:      clock,
		scheduler:  reviewScheduler,
	}, nil
}

func (u *PreviewReviewSchedule) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawReviewItemID string,
	performedAt time.Time,
) ([]domain.SchedulePreview, error) {
	if userID.IsZero() {
		return nil, errors.New("preview review schedule: user id is required")
	}
	reviewItemID, err := domain.ParseReviewItemID(rawReviewItemID)
	if err != nil {
		return nil, fmt.Errorf("preview review schedule: %w", err)
	}
	item, err := u.repository.FindReviewItemByID(ctx, userID, reviewItemID)
	if err != nil {
		return nil, fmt.Errorf("preview review schedule: %w", err)
	}
	location, _, err := loadLocation(ctx, u.timeZones, userID)
	if err != nil {
		return nil, fmt.Errorf("preview review schedule: %w", err)
	}
	previews, err := item.PreviewSchedule(performedAt, u.clock(), location, u.scheduler)
	if err != nil {
		return nil, fmt.Errorf("preview review schedule: %w", err)
	}
	return previews, nil
}
