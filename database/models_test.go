package database

import (
	"context"
	"time"
)

//go:generate neurogonesis models methods --format=goimports --single-file --type=SoftDeletable,SoftDeletableNoHooks,Updateable .

type TestModel struct {
	ID             int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	FieldSetBefore string
	FieldSetAfter  string
	Integer        int
}

func (s *TestModel) BeforeDelete(_ context.Context, _ DB) error {
	s.FieldSetBefore = "before"
	return nil
}

func (s *TestModel) AfterDelete(_ context.Context, _ DB) error {
	s.FieldSetAfter = "after"
	return nil
}

func (s *TestModel) BeforeUpdate(_ context.Context, _ DB) error {
	s.FieldSetBefore = "before"
	return nil
}

func (s *TestModel) AfterUpdate(_ context.Context, _ DB) error {
	s.FieldSetAfter = "after"
	return nil
}

func (s *TestModel) BeforeInsert(_ context.Context, _ DB) error {
	if s.FieldSetBefore == "" {
		s.FieldSetBefore = "before"
	}
	return nil
}

func (s *TestModel) AfterInsert(_ context.Context, _ DB) error {
	s.FieldSetAfter = "after"
	return nil
}

type SoftDeletableNoHooks struct {
	ID        int
	DeletedAt *time.Time
	Integer   int
}

type Updateable struct {
	ID        int
	UpdatedAt time.Time
	Integer   int
}
