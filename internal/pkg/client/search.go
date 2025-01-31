package client

import (
	"context"
	"fmt"
	"strings"

	options "github.com/breathbath/go_utils/v2/pkg/config"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/models"
)

type DataProvider interface {
	GetClients(ctx context.Context) (cls []*models.Client, err error)
}

type Search struct {
	DataProvider DataProvider
}

func (s *Search) Search(ctx context.Context, term string, params *options.ParameterBag) (foundCls []*models.Client, err error) {
	cls, err := s.DataProvider.GetClients(ctx)
	if err != nil {
		return foundCls, err
	}

	foundCls = s.findInClientsList(cls, term)
	return
}

func (s *Search) FindOne(ctx context.Context, searchTerm string, params *options.ParameterBag) (*models.Client, error) {
	clients, err := s.Search(ctx, searchTerm, params)
	if err != nil {
		return &models.Client{}, err
	}

	if len(clients) == 0 {
		return &models.Client{}, fmt.Errorf("unknown client '%s'", searchTerm)
	}

	if len(clients) == 1 {
		return clients[0], nil
	}

	return &models.Client{}, fmt.Errorf("client identified by '%s' is ambiguous, use a more precise name or use the client id", searchTerm)
}

func (s *Search) findInClientsList(cls []*models.Client, term string) (foundCls []*models.Client) {
	terms := strings.Split(term, ",")
	for i := range terms {
		terms[i] = strings.ToLower(terms[i])
	}

	foundCls = make([]*models.Client, 0)
	for i := range cls {
		cl := cls[i]
		curClientName := strings.ToLower(cl.Name)
		curClientID := strings.ToLower(cl.ID)

		for i := range terms {
			curTerm := terms[i]
			if strings.HasPrefix(curClientName, curTerm) {
				foundCls = append(foundCls, cl)
			} else if strings.HasPrefix(curClientID, curTerm) {
				foundCls = append(foundCls, cl)
			}
		}
	}

	return
}
