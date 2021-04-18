package communication

import "github.com/wufe/polo/pkg/utils"

type PubSubBuilder struct {
	mutexBuilder utils.MutexBuilder
}

func NewPubSubBuilder(mutexBuilder utils.MutexBuilder) *PubSubBuilder {
	return &PubSubBuilder{
		mutexBuilder: mutexBuilder,
	}
}

func (b *PubSubBuilder) Build() *PubSub {
	return newPubSub(b.mutexBuilder)
}
