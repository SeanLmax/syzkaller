// Copyright 2025 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/syzkaller/syz-cluster/pkg/api"
	"github.com/google/syzkaller/syz-cluster/pkg/app"
	"github.com/google/syzkaller/syz-cluster/pkg/blob"
	"github.com/google/syzkaller/syz-cluster/pkg/db"
)

type FindingService struct {
	findingRepo *db.FindingRepository
	blobStorage blob.Storage
}

func NewFindingService(env *app.AppEnvironment) *FindingService {
	return &FindingService{
		findingRepo: db.NewFindingRepository(env.Spanner),
		blobStorage: env.BlobStorage,
	}
}

func (s *FindingService) Save(ctx context.Context, req *api.NewFinding) error {
	var reportURI, logURI string
	var err error
	if len(req.Log) > 0 {
		logURI, err = s.blobStorage.Store(bytes.NewReader(req.Log))
		if err != nil {
			return fmt.Errorf("failed to save the log: %w", err)
		}
	}
	if len(req.Report) > 0 {
		reportURI, err = s.blobStorage.Store(bytes.NewReader(req.Report))
		if err != nil {
			return fmt.Errorf("failed to save the report: %w", err)
		}
	}
	// TODO: if it's not actually addded, the blob records will be orphaned.
	err = s.findingRepo.Save(ctx, &db.Finding{
		SessionID: req.SessionID,
		TestName:  req.TestName,
		Title:     req.Title,
		ReportURI: reportURI,
		LogURI:    logURI,
	})
	if err == db.ErrFindingExists {
		// It's ok, just ignore.
		return nil
	}
	return err
}
