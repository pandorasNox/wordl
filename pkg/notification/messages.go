package notification

import "slices"

type message string

type notifier struct {
	errorMsgs	[]message
	infoMsgs	[]message
	successMsgs	[]message
}

func NewNotifier() notifier {
	return notifier{}
}

type TemplateDataMessages struct {
	ErrMsgs     []message
	InfoMsgs    []message
	SuccessMsgs []message
}

func (n *notifier) AddError(msg string) {
	n.errorMsgs = append(n.errorMsgs, message(msg))
}

func (n *notifier) AddInfo(msg string) {
	n.infoMsgs = append(n.infoMsgs, message(msg))
}

func (n *notifier) AddSuccess(msg string) {
	n.successMsgs = append(n.successMsgs, message(msg))
}

func (n *notifier) ToTemplate() TemplateDataMessages {
	return TemplateDataMessages{
		slices.Clone(n.errorMsgs),
		slices.Clone(n.infoMsgs),
		slices.Clone(n.successMsgs),
	}
}