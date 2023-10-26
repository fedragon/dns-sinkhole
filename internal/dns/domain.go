package dns

import (
	"errors"
	"fmt"
	"strings"
)

type Domain struct {
	name     string
	parent   *Domain
	children map[string]*Domain
}

func NewDomain(name string) (*Domain, error) {
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		return nil, fmt.Errorf("invalid domain: %s", name)
	}

	root := &Domain{name: name[idx+1:]}
	if err := upsert(root, name[:idx]); err != nil {
		return nil, err
	}

	return root, nil
}

func (d *Domain) Register(subdomain string) error {
	idx := strings.LastIndex(subdomain, ".")
	if idx == -1 {
		return fmt.Errorf("invalid domain: %s", subdomain)
	}

	if d.name != subdomain[idx+1:] {
		return fmt.Errorf("%s is not a subdomain of %s", subdomain, d.name)
	}

	return upsert(d, subdomain[:idx])
}

func (d *Domain) Contains(name string) bool {
	if len(d.children) == 0 {
		return name == d.name
	}

	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		if d.name != name {
			return false
		}

		if d.name != "www" && len(d.children) > 0 {
			return d.children["@"] != nil
		}

		return false
	}

	part := parts[len(parts)-1]
	if d.name != part {
		return false
	}

	for _, child := range d.children {
		if child.Contains(strings.Join(parts[:len(parts)-1], ".")) {
			return true
		}
	}

	return false
}

func upsert(parent *Domain, name string) error {
	if name == "" {
		return errors.New("domain cannot be empty")
	}

	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		if parent == nil {
			return errors.New("invalid domain")
		}

		if len(parent.children) == 0 {
			parent.children = make(map[string]*Domain)
		}

		node, ok := parent.children[name]
		if !ok {
			node = &Domain{name: name, parent: parent}
			parent.children[name] = node
		}

		if name != "www" {
			if len(node.children) == 0 {
				node.children = make(map[string]*Domain)
			}

			if _, ok := node.children["@"]; !ok {
				node.children["@"] = &Domain{name: "@", parent: node}
			}
		}

		return nil
	}

	part := name[idx+1:]
	if parent == nil {
		return upsert(&Domain{name: part}, name[:idx])
	}

	if len(parent.children) == 0 {
		parent.children = make(map[string]*Domain)
	}

	node, ok := parent.children[part]
	if !ok {
		node = &Domain{name: part, parent: parent}
		parent.children[part] = node
	}

	return upsert(node, name[:idx])
}
