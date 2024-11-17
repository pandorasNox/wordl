package notification

import "slices"

type message string

type Notifier struct {
	errorMsgs   []message
	infoMsgs    []message
	successMsgs []message
}

func NewNotifier() Notifier {
	return Notifier{}
}

type TemplateDataMessages struct {
	ErrMsgs     []message
	InfoMsgs    []message
	SuccessMsgs []message
}

func (n *Notifier) AddError(msg string) {
	n.errorMsgs = append(n.errorMsgs, message(msg))
}

func (n *Notifier) AddInfo(msg string) {
	n.infoMsgs = append(n.infoMsgs, message(msg))
}

func (n *Notifier) AddSuccess(msg string) {
	n.successMsgs = append(n.successMsgs, message(msg))
}

func (n *Notifier) ToTemplate() TemplateDataMessages {
	return TemplateDataMessages{
		slices.Clone(n.errorMsgs),
		slices.Clone(n.infoMsgs),
		slices.Clone(n.successMsgs),
	}
}
