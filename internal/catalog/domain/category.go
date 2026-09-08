package domain

import "github.com/google/uuid"

type Category struct {
	id         uuid.UUID
	name, slug string
}

func ReconstructCategory(id uuid.UUID, name, slug string) *Category { return &Category{id: id, name: name, slug: slug} }
func (c *Category) ID() uuid.UUID { return c.id }
func (c *Category) Name() string  { return c.name }
func (c *Category) Slug() string  { return c.slug }
