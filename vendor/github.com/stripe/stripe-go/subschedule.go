package stripe

import (
	"encoding/json"
)

// SubscriptionScheduleStatus is the list of allowed values for the schedule's status.
type SubscriptionScheduleStatus string

// List of values that SubscriptionScheduleStatus can take.
const (
	SubscriptionScheduleStatusActive    SubscriptionScheduleStatus = "active"
	SubscriptionScheduleStatusCanceled  SubscriptionScheduleStatus = "canceled"
	SubscriptionScheduleStatusCompleted SubscriptionScheduleStatus = "completed"
	SubscriptionScheduleStatusPastDue   SubscriptionScheduleStatus = "not_started"
	SubscriptionScheduleStatusTrialing  SubscriptionScheduleStatus = "released"
)

// SubscriptionScheduleRenewalBehavior describe what happens to a schedule when it ends.
type SubscriptionScheduleRenewalBehavior string

// List of values that SubscriptionScheduleRenewalBehavior can take.
const (
	SubscriptionScheduleRenewalBehaviorNone    SubscriptionScheduleRenewalBehavior = "none"
	SubscriptionScheduleRenewalBehaviorRenew   SubscriptionScheduleRenewalBehavior = "release"
	SubscriptionScheduleRenewalBehaviorRelease SubscriptionScheduleRenewalBehavior = "renew"
)

// SubscriptionScheduleInvoiceSettingsParams is a structure representing the parameters allowed to
// control invoice settings on invoices associated with a subscription schedule.
type SubscriptionScheduleInvoiceSettingsParams struct {
	DaysUntilDue *int64 `form:"days_until_due"`
}

// SubscriptionSchedulePhaseItemParams is a structure representing the parameters allowed to control
// a specic plan on a phase on a subscription schedule.
type SubscriptionSchedulePhaseItemParams struct {
	BillingThresholds *SubscriptionItemBillingThresholdsParams `form:"billing_thresholds"`
	Plan              *string                                  `form:"plan"`
	Quantity          *int64                                   `form:"quantity"`
}

// SubscriptionSchedulePhaseParams is a structure representing the parameters allowed to control
// a phase on a subscription schedule.
type SubscriptionSchedulePhaseParams struct {
	ApplicationFeePercent *int64                                 `form:"application_fee_percent"`
	Coupon                *string                                `form:"coupon"`
	EndDate               *int64                                 `form:"end_date"`
	Iterations            *int64                                 `form:"iterations"`
	Plans                 []*SubscriptionSchedulePhaseItemParams `form:"plans"`
	Trial                 *bool                                  `form:"trial"`
	TrialEnd              *int64                                 `form:"trial_end"`
}

// SubscriptionScheduleParams is the set of parameters that can be used when creating or updating a
// subscription schedule.
type SubscriptionScheduleParams struct {
	Params            `form:"*"`
	Billing           *string                                    `form:"billing"`
	BillingThresholds *SubscriptionBillingThresholdsParams       `form:"billing_thresholds"`
	Customer          *string                                    `form:"customer"`
	FromSubscription  *string                                    `form:"from_subscription"`
	InvoiceSettings   *SubscriptionScheduleInvoiceSettingsParams `form:"invoice_settings"`
	Phases            []*SubscriptionSchedulePhaseParams         `form:"phases"`
	RenewalBehavior   *string                                    `form:"renewal_behavior"`
	StartDate         *int64                                     `form:"start_date"`
}

// SubscriptionScheduleCancelParams is the set of parameters that can be used when canceling a
// subscription schedule.
type SubscriptionScheduleCancelParams struct {
	Params     `form:"*"`
	InvoiceNow *bool `form:"invoice_now"`
	Prorate    *bool `form:"prorate"`
}

// SubscriptionScheduleReleaseParams is the set of parameters that can be used when releasing a
// subscription schedule.
type SubscriptionScheduleReleaseParams struct {
	Params             `form:"*"`
	PreserveCancelDate *bool `form:"preserve_cancel_date"`
}

// SubscriptionScheduleListParams is the set of parameters that can be used when listing
// subscription schedules.
type SubscriptionScheduleListParams struct {
	ListParams       `form:"*"`
	CanceledAt       int64             `form:"canceled_at"`
	CanceledAtRange  *RangeQueryParams `form:"canceled_at"`
	CompletedAt      int64             `form:"completed_at"`
	CompletedAtRange *RangeQueryParams `form:"completed_at"`
	Created          int64             `form:"created"`
	CreatedRange     *RangeQueryParams `form:"created"`
	Customer         string            `form:"customer"`
	ReleasedAt       int64             `form:"released_at"`
	ReleasedAtRange  *RangeQueryParams `form:"released_at"`
	Scheduled        *bool             `form:"scheduled"`
}

// SubscriptionScheduleCurrentPhase is a structure the current phase's period.
type SubscriptionScheduleCurrentPhase struct {
	EndDate   int64 `json:"end_date"`
	StartDate int64 `json:"start_date"`
}

// SubscriptionScheduleInvoiceSettings is a structure representing the settings applied to invoices
// associated with a subscription schedule.
type SubscriptionScheduleInvoiceSettings struct {
	DaysUntilDue int64 `json:"days_until_due"`
}

// SubscriptionSchedulePhaseItem represents plan details for a given phase
type SubscriptionSchedulePhaseItem struct {
	BillingThresholds *SubscriptionItemBillingThresholds `json:"billing_thresholds"`
	Plan              *Plan                              `json:"plan"`
	Quantity          int64                              `json:"quantity"`
}

// SubscriptionSchedulePhase is a structure a phase of a subscription schedule.
type SubscriptionSchedulePhase struct {
	ApplicationFeePercent float64                          `json:"application_fee_percent"`
	Coupon                *Coupon                          `json:"coupon"`
	EndDate               int64                            `json:"end_date"`
	Plans                 []*SubscriptionSchedulePhaseItem `json:"plans"`
	StartDate             int64                            `json:"start_date"`
	TaxPercent            float64                          `json:"tax_percent"`
	TrialEnd              int64                            `json:"trial_end"`
}

// SubscriptionScheduleRenewalInterval represents the interval and duration of a schedule.
type SubscriptionScheduleRenewalInterval struct {
	Interval PlanInterval `form:"interval"`
	Length   int64        `form:"length"`
}

// SubscriptionSchedule is the resource representing a Stripe subscription schedule.
type SubscriptionSchedule struct {
	Billing              SubscriptionBilling                  `json:"billing"`
	BillingThresholds    *SubscriptionBillingThresholds       `json:"billing_thresholds"`
	CanceledAt           int64                                `json:"canceled_at"`
	CompletedAt          int64                                `json:"completed_at"`
	Created              int64                                `json:"created"`
	CurrentPhase         *SubscriptionScheduleCurrentPhase    `json:"current_phase"`
	Customer             *Customer                            `json:"customer"`
	ID                   string                               `json:"id"`
	InvoiceSettings      *SubscriptionScheduleInvoiceSettings `json:"invoice_settings"`
	Livemode             bool                                 `json:"livemode"`
	Metadata             map[string]string                    `json:"metadata"`
	Object               string                               `json:"object"`
	Phases               []*SubscriptionSchedulePhase         `json:"phases"`
	ReleasedSubscription *Subscription                        `json:"released_subscription"`
	Revision             *SubscriptionScheduleRevision        `json:"revision"`
	RenewalBehavior      SubscriptionScheduleRenewalBehavior  `json:"renewal_behavior"`
	RenewalInterval      *SubscriptionScheduleRenewalInterval `json:"renewal_interval"`
	Status               SubscriptionScheduleStatus           `json:"status"`
	Subscription         *Subscription                        `json:"subscription"`
}

// SubscriptionScheduleList is a list object for subscription schedules.
type SubscriptionScheduleList struct {
	ListMeta
	Data []*SubscriptionSchedule `json:"data"`
}

// UnmarshalJSON handles deserialization of a SubscriptionSchedule.
// This custom unmarshaling is needed because the resulting
// property may be an id or the full struct if it was expanded.
func (s *SubscriptionSchedule) UnmarshalJSON(data []byte) error {
	if id, ok := ParseID(data); ok {
		s.ID = id
		return nil
	}

	type schedule SubscriptionSchedule
	var v schedule
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*s = SubscriptionSchedule(v)
	return nil
}
