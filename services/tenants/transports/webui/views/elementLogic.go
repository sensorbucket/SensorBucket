package views

import (
	ory "github.com/ory/client-go"
	"github.com/samber/lo"
)

func isMFA(flow *ory.LoginFlow) bool {
	nodes := flow.Ui.GetNodes()
	return isLoggedIn(flow) && (hasGroup(nodes, "totp") ||
		hasGroup(nodes, "webauthn") ||
		hasGroup(nodes, "lookup_secret"))
}

func isLoggedIn(flow *ory.LoginFlow) bool {
	return flow.GetRequestedAal() == "aal2" || flow.GetRefresh()
}

func hasGroup(n []ory.UiNode, group string) bool {
	return lo.ContainsBy(n, func(item ory.UiNode) bool {
		return item.Group == group
	})
}

//
// Node filtering
//

type predicate func(ory.UiNode) bool

type filterChain struct {
	nodes      []ory.UiNode
	predicates []predicate
}

func filter(n []ory.UiNode) filterChain {
	return filterChain{
		nodes:      n,
		predicates: make([]predicate, 0),
	}
}

func (fc filterChain) Group(g string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.GetGroup() == g
	}
	fc.predicates = append(fc.predicates, p)
	return fc
}

func (fc filterChain) InputType(t string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.Attributes.UiNodeInputAttributes != nil && n.Attributes.UiNodeInputAttributes.GetType() == t
	}
	fc.predicates = append(fc.predicates, p)
	return fc
}

func (fc filterChain) InputName(name string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.Attributes.UiNodeInputAttributes != nil && n.Attributes.UiNodeInputAttributes.GetName() == name
	}
	fc.predicates = append(fc.predicates, p)
	return fc
}

func (fc filterChain) GetWithThese() []ory.UiNode {
	for _, p := range fc.predicates {
		fc.nodes = lo.Filter(fc.nodes, func(item ory.UiNode, _ int) bool {
			return p(item)
		})
	}
	return fc.nodes
}

func (fc filterChain) GetWithoutThese() []ory.UiNode {
	applicables := []ory.UiNode{}
	for _, p := range fc.predicates {
		fc.nodes = lo.Filter(fc.nodes, func(item ory.UiNode, _ int) bool {
			if p(item) {
				return true
			}
			applicables = append(applicables, item)
			return false
		})
	}
	return applicables
}

func (fc filterChain) ContinueWithThese() filterChain {
	return filter(fc.GetWithThese())
}

func (fc filterChain) ContinueWithoutThese() filterChain {
	return filter(fc.GetWithoutThese())
}
