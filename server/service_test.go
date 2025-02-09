// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package server

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/moov-io/ach"
)

// test mocks are in mock_test.go

// CreateFile tests
func TestCreateFile(t *testing.T) {
	s := mockServiceInMemory()
	id, err := s.CreateFile(mockFileHeader())
	if err != nil {
		t.Fatal(err.Error())
	}
	if id != "12345" {
		t.Errorf("expected %s received %s w/ error %s", "12345", id, err)
	}
}
func TestCreateFileIDExists(t *testing.T) {
	s := mockServiceInMemory()
	h := ach.FileHeader{ID: "98765"}
	id, err := s.CreateFile(&h)
	if err != ErrAlreadyExists {
		t.Errorf("expected %s received %s w/ error %s", "ErrAlreadyExists", id, err)
	}
}

func TestCreateFileNoID(t *testing.T) {
	s := mockServiceInMemory()
	h := ach.NewFileHeader()
	id, err := s.CreateFile(&h)
	if len(id) < 3 {
		t.Errorf("expected %s received %s w/ error %s", "NextID", id, err)
	}
	if err != nil {
		t.Fatal(err.Error())
	}
}

// Service.GetFile tests

func TestGetFile(t *testing.T) {
	s := mockServiceInMemory()
	f, err := s.GetFile("98765")
	if err != nil {
		t.Errorf("expected %s received %s w/ error %s", "98765", f.ID, err)
	}
}

func TestGetFileNotFound(t *testing.T) {
	s := mockServiceInMemory()
	f, err := s.GetFile("12345")
	if err != ErrNotFound {
		t.Errorf("expected %s received %s w/ error %s", "ErrNotFound", f.ID, err)
	}
}

// Service.GetFiles tests

func TestGetFiles(t *testing.T) {
	s := mockServiceInMemory()
	files := s.GetFiles()
	if len(files) != 1 {
		t.Errorf("expected %s received %v", "1", len(files))
	}
}

// Service.DeleteFile tests

func TestDeleteFile(t *testing.T) {
	s := mockServiceInMemory()
	err := s.DeleteFile("98765")
	if err != nil {
		t.Errorf("expected %s received %s", "nil", err)
	}
	_, err = s.GetFile("98765")
	if err != ErrNotFound {
		t.Errorf("expected %s received %s", "ErrNotFound", err)
	}
}

// Service.GetFileContents tests

func TestGetFileContents(t *testing.T) {
	s := mockServiceInMemory()
	id, err := s.CreateFile(mockFileHeader())
	if err != nil {
		t.Fatal(err.Error())
	}

	// make the file valid
	batch := mockBatchWEB()
	s.CreateBatch(id, batch)

	// build file
	r, err := s.GetFileContents(id)
	if err != nil {
		if !strings.Contains(err.Error(), "mandatory ") {
			t.Fatal(err.Error())
		}
	}
	if r != nil {
		bs, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal(err.Error())
		}

		if len(bs) == 0 {
			t.Fatal("expected to read fil")
		}
	}
}

// Service.ValidateFile tests

func TestValidateFile(t *testing.T) {
	s := mockServiceInMemory()
	id, err := s.CreateFile(mockFileHeader())
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := s.ValidateFile(id); err != nil {
		if !strings.Contains(err.Error(), "mandatory ") {
			t.Fatal(err.Error())
		}
	}
}

func TestValidateFileMissing(t *testing.T) {
	s := mockServiceInMemory()
	err := s.ValidateFile("missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateFileBad(t *testing.T) {
	s := mockServiceInMemory()

	fId, _ := s.CreateFile(mockFileHeader())

	// setup batch
	bh := mockBatchHeaderWeb()
	bh.ID = "11111"
	b, _ := ach.NewBatch(bh)
	bId, e1 := s.CreateBatch(fId, b)
	batch, e2 := s.GetBatch(fId, bId)
	if batch == nil {
		t.Fatalf("couldn't get batch, e1=%v, e2=%v", e1, e2)
	}

	// setup file, add batch
	f, err := s.GetFile(fId)
	if f == nil {
		t.Fatalf("couldn't get file: %v", err)
	}
	if len(f.AddBatch(batch)) == 0 {
		t.Fatal("problem adding batch to file")
	}

	// validate
	if err := s.ValidateFile(fId); err == nil {
		t.Fatal("expected error")
	}
}

// Service.CreateBatch tests

// TestCreateBatch tests creating a new batch when file.ID exists and batch.id does not exist
func TestCreateBatch(t *testing.T) {
	s := mockServiceInMemory()
	bh := mockBatchHeaderWeb()
	bh.ID = "11111"
	b, _ := ach.NewBatch(bh)
	id, err := s.CreateBatch("98765", b)
	if err != nil {
		t.Fatal(err.Error())
	}
	if id != "11111" {
		t.Errorf("expected %s received %s w/ error %v", "11111", id, err)
	}
}

// TestCreateBatchIDExists Create a new batch with batch.id already present. Should fail.
func TestCreateBatchIDExists(t *testing.T) {
	s := mockServiceInMemory()
	b, _ := ach.NewBatch(mockBatchHeaderWeb())
	id, err := s.CreateBatch("98765", b)
	if err != ErrAlreadyExists {
		t.Errorf("expected %s received %s w/ error %v", "ErrAlreadyExists", id, err)
	}
}

// TestCreateBatchFileIDExits create a batch when the file.id does not exist. Should fail.
func TestCreateBatchFileIDExits(t *testing.T) {
	s := mockServiceInMemory()
	b, _ := ach.NewBatch(mockBatchHeaderWeb())
	id, err := s.CreateBatch("55555", b)
	if err != ErrNotFound {
		t.Errorf("expected %s received %s w/ error %v", "ErrNotFound", id, err)
	}
}

// TestCreateBatchIDBank create a new batch when the batch.id is nil but file.id is valid. Should generate batch.id and save.
func TestCreateBatchIDBlank(t *testing.T) {
	s := mockServiceInMemory()
	bh := mockBatchHeaderWeb()
	bh.ID = ""
	b, _ := ach.NewBatch(bh)
	id, err := s.CreateBatch("98765", b)
	if len(id) < 3 {
		t.Errorf("expected %s received %s w/ error %v", "NextID", id, err)
	}
	if err != nil {
		t.Fatal(err.Error())
	}
}

// Service.GetBatch

// TestGetBatch return a batch for the existing file.id and batch.id
func TestGetBatch(t *testing.T) {
	s := mockServiceInMemory()
	b, err := s.GetBatch("98765", "54321")
	if err != nil {
		t.Errorf("problem getting batch: %v", err)
	}
	if b.ID() != "54321" {
		t.Errorf("expected %s received %s w/ error %v", "54321", b.ID(), err)
	}
}

// TestGetBatchNotFound return a failure if the batch.id is not found
func TestGetBatchNotFound(t *testing.T) {
	s := mockServiceInMemory()
	b, err := s.GetBatch("98765", "55555")
	if err != ErrNotFound {
		t.Errorf("expected %s received %#v w/ error %v", "ErrNotFound", b, err)
	}
}

// Service.GetBatches

// TestGetBatches return a list of batches for the supplied file.id
func TestGetBatches(t *testing.T) {
	s := mockServiceInMemory()
	batches := s.GetBatches("98765")
	if len(batches) != 1 {
		t.Errorf("expected %s received %v", "1", len(batches))
	}
}

// Service.DeleteBatch

// TestDeleteBatch removes a batch with existing file and batch id.
func TestDeleteBatch(t *testing.T) {
	s := mockServiceInMemory()
	err := s.DeleteBatch("98765", "54321")
	if err != nil {
		t.Errorf("expected %s received error %v", "nil", err)
	}
}
