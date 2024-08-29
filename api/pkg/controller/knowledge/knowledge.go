package knowledge

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/helixml/helix/api/pkg/config"
	"github.com/helixml/helix/api/pkg/extract"
	"github.com/helixml/helix/api/pkg/rag"
	"github.com/helixml/helix/api/pkg/store"
	"github.com/helixml/helix/api/pkg/types"

	"github.com/rs/zerolog/log"
)

type Reconciler struct {
	config       *config.ServerConfig
	store        store.Store
	extractor    extract.Extractor // Unstructured.io or equivalent
	httpClient   *http.Client
	ragClient    rag.RAG                                   // Default server RAG client
	newRagClient func(settings *types.RAGSettings) rag.RAG // Custom RAG server client constructor
	wg           sync.WaitGroup
}

func New(config *config.ServerConfig, store store.Store, extractor extract.Extractor, ragClient rag.RAG) *Reconciler {
	return &Reconciler{
		config:     config,
		store:      store,
		extractor:  extractor,
		httpClient: http.DefaultClient,
		ragClient:  ragClient,
		newRagClient: func(settings *types.RAGSettings) rag.RAG {
			return rag.NewLlamaindex(settings)
		},
	}
}

func (r *Reconciler) Start(ctx context.Context) error {
	err := r.reset(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("knowledge state reset failed")
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		r.runIndexer(ctx)
	}()

	wg.Add(1)
	go func() {
		r.startCron(ctx)
	}()

	wg.Add(1)
	go func() {
		r.runCronManager(ctx)
	}()

	wg.Wait()

	return nil
}

func (r *Reconciler) runIndexer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			err := r.index(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("failed to index knowledge")
			}
		}
	}
}

// runCronManager is responsible for reconciling the cron jobs in the database
// with the actual cron jobs that are running.
func (r *Reconciler) runCronManager(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			err := r.reconcileCronJobs(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("failed to reconcile cron jobs")
			}
		}
	}
}

func (r *Reconciler) reset(ctx context.Context) error {
	data, err := r.store.ListKnowledge(ctx, &store.ListKnowledgeQuery{
		State: types.KnowledgeStateIndexing,
	})
	if err != nil {
		return fmt.Errorf("failed to get knowledge entries, error: %w", err)
	}

	for _, k := range data {
		k.State = types.KnowledgeStatePending

		_, err = r.store.UpdateKnowledge(ctx, k)
		if err != nil {
			log.Error().Err(err).Msg("failed to reset knowledge back into pending during reset")
		}
	}

	return nil
}
