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

type filterChain []predicate

var filter filterChain

func (fc filterChain) Group(g string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.GetGroup() == g
	}
	fc = append(fc, p)
	return fc
}

func (fc filterChain) InputType(t string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.Attributes.UiNodeInputAttributes != nil && n.Attributes.UiNodeInputAttributes.GetType() == t
	}
	fc = append(fc, p)
	return fc
}

func (fc filterChain) InputName(name string) filterChain {
	p := func(n ory.UiNode) bool {
		return n.Attributes.UiNodeInputAttributes != nil && n.Attributes.UiNodeInputAttributes.GetName() == name
	}
	fc = append(fc, p)
	return fc
}

func (fc filterChain) GetWithThese(n []ory.UiNode) []ory.UiNode {
	for _, p := range fc {
		n = lo.Filter(n, func(item ory.UiNode, _ int) bool {
			return p(item)
		})
	}
	return n
}

func (fc filterChain) GetWithoutThese(n []ory.UiNode) []ory.UiNode {
	applicables := []ory.UiNode{}
	for _, p := range fc {
		n = lo.Filter(n, func(item ory.UiNode, _ int) bool {
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
	p := func(item ory.UiNode) bool {
		return len(fc.GetWithThese([]ory.UiNode{item})) > 0
	}
	return filterChain{p}
}

func (fc filterChain) ContinueWithoutThese() filterChain {
	p := func(item ory.UiNode) bool {
		return len(fc.GetWithoutThese([]ory.UiNode{item})) > 0
	}
	return filterChain{p}
}
