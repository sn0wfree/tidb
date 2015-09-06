// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package boltdb

import (
	"os"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/store/localstore/engine"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testSuite{})

type testSuite struct {
	db engine.DB
}

const testPath = "/tmp/test-tidb-boltdb"

func (s *testSuite) SetUpSuite(c *C) {
	var (
		d   Driver
		err error
	)
	s.db, err = d.Open(testPath)
	c.Assert(err, IsNil)
}

func (s *testSuite) TearDownSuite(c *C) {
	s.db.Close()
	os.Remove(testPath)
}

func (s *testSuite) TestDB(c *C) {
	db := s.db

	b := db.NewBatch()
	b.Put([]byte("a"), []byte("1"))
	b.Put([]byte("b"), []byte("2"))
	b.Delete([]byte("c"))

	err := db.Commit(b)
	c.Assert(err, IsNil)

	v, err := db.Get([]byte("c"))
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)

	v, err = db.Get([]byte("a"))
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, []byte("1"))

	snap, err := db.GetSnapshot()
	c.Assert(err, IsNil)

	v, err = snap.Get([]byte("a"))
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, []byte("1"))

	v, err = snap.Get([]byte("c"))
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)

	iter := snap.NewIterator(nil)
	c.Assert(iter.Next(), Equals, true)
	c.Assert(iter.Key(), DeepEquals, []byte("a"))
	c.Assert(iter.Next(), Equals, true)
	c.Assert(iter.Key(), DeepEquals, []byte("b"))
	c.Assert(iter.Next(), Equals, false)
	iter.Release()

	iter = snap.NewIterator([]byte("b"))
	c.Assert(iter.Next(), Equals, true)
	c.Assert(iter.Key(), DeepEquals, []byte("b"))
	c.Assert(iter.Value(), DeepEquals, []byte("2"))
	c.Assert(iter.Next(), Equals, false)

	snap.Release()
}
