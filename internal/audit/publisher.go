package audit

import (
	"sync"
)

var _ Publisher = &Notifier{}

type Notifier struct {
	audits map[string]Audit
	wg     sync.WaitGroup
}

func NewNotifier() *Notifier {
	return &Notifier{
		audits: make(map[string]Audit),
	}
}

func (n *Notifier) Register(a Audit) {
	if n.audits == nil {
		n.audits = make(map[string]Audit)
	}

	n.audits[a.GetID()] = a
}

func (n *Notifier) Notify(event Event) {
	if n == nil {
		return
	}
	for _, audit := range n.audits {
		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			audit.Update(event)
		}()
	}
}

func (n *Notifier) Wait() {
	n.wg.Wait()
}
