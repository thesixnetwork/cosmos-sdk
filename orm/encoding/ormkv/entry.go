package ormkv

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// Entry defines a logical representation of a kv-store entry for ORM instances.
type Entry interface {
	fmt.Stringer

	// GetTableName returns the table-name (equivalent to the fully-qualified
	// proto message name) this entry corresponds to.
	GetTableName() protoreflect.FullName

	doNotImplement()
}

// PrimaryKeyEntry represents a logically decoded primary-key entry.
type PrimaryKeyEntry struct {
	// Key represents the primary key values.
	Key []protoreflect.Value

	// Value represents the message stored under the primary key.
	Value proto.Message
}

func (p *PrimaryKeyEntry) GetTableName() protoreflect.FullName {
	return p.Value.ProtoReflect().Descriptor().FullName()
}

func (p *PrimaryKeyEntry) String() string {
	msg := p.Value
	name := msg.ProtoReflect().Descriptor().FullName()
	msgBz, err := protojson.Marshal(msg)
	msgStr := string(msgBz)
	if err != nil {
		msgStr = fmt.Sprintf("%s:%+v", name, msg)
	}
	return fmt.Sprintf("PK:%s:%s:%s", name, fmtValues(p.Key), msgStr)
}

func fmtValues(values []protoreflect.Value) string {
	var xs []interface{}
	for _, v := range values {
		xs = append(xs, v.Interface())
	}
	list, err := structpb.NewList(xs)
	if err != nil {
		return fmt.Sprintf("%+v", values)
	}
	bz, err := protojson.Marshal(list)
	if err != nil {
		return fmt.Sprintf("%+v", values)
	}
	return string(bz)
}

func (p *PrimaryKeyEntry) doNotImplement() {}

// IndexKeyEntry represents a logically decoded index entry.
type IndexKeyEntry struct {

	// TableName is the table this entry represents.
	TableName protoreflect.FullName

	// Fields are the index fields this entry represents.
	Fields []protoreflect.Name

	// IsUnique indicates whether this index is unique or not.
	IsUnique bool

	// IndexValues represent the index values.
	IndexValues []protoreflect.Value

	// PrimaryKey represents the primary key values, it is empty if this is a
	// prefix key
	PrimaryKey []protoreflect.Value
}

func (i *IndexKeyEntry) GetTableName() protoreflect.FullName {
	return i.TableName
}

func (i *IndexKeyEntry) doNotImplement() {}

func (i *IndexKeyEntry) string() string {
	return fmt.Sprintf("%s%s:%s:%s", i.TableName, i.Fields, fmtValues(i.IndexValues), fmtValues(i.PrimaryKey))
}

func (i *IndexKeyEntry) String() string {
	if i.IsUnique {
		return fmt.Sprintf("UNIQ:%s", i.string())
	} else {

		return fmt.Sprintf("IDX:%s", i.string())
	}
}

// SeqEntry represents a sequence for tables with auto-incrementing primary keys.
type SeqEntry struct {

	// TableName is the table this entry represents.
	TableName protoreflect.FullName

	// Value is the uint64 value stored for this sequence.
	Value uint64
}

func (s *SeqEntry) GetTableName() protoreflect.FullName {
	return s.TableName
}

func (s *SeqEntry) doNotImplement() {}

func (s *SeqEntry) String() string {
	return fmt.Sprintf("SEQ:%s:%d", s.TableName, s.Value)
}

var _, _, _ Entry = &PrimaryKeyEntry{}, &IndexKeyEntry{}, &SeqEntry{}
