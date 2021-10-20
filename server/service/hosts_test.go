package service

import (
	"context"
	"testing"
	"time"

	"github.com/fleetdm/fleet/v4/server/authz"
	"github.com/fleetdm/fleet/v4/server/fleet"
	"github.com/fleetdm/fleet/v4/server/mock"
	"github.com/fleetdm/fleet/v4/server/test"
	"github.com/stretchr/testify/require"
)

func TestListHosts(t *testing.T) {
	ds := new(mock.Store)
	svc := newTestService(ds, nil, nil)

	ds.ListHostsFunc = func(ctx context.Context, filter fleet.TeamFilter, opt fleet.HostListOptions) ([]*fleet.Host, error) {
		return []*fleet.Host{
			{ID: 1},
		}, nil
	}

	hosts, err := svc.ListHosts(test.UserContext(test.UserAdmin), fleet.HostListOptions{})
	require.NoError(t, err)
	require.Len(t, hosts, 1)

	// anyone can list hosts
	hosts, err = svc.ListHosts(test.UserContext(test.UserNoRoles), fleet.HostListOptions{})
	require.NoError(t, err)
	require.Len(t, hosts, 1)

	// a user is required
	_, err = svc.ListHosts(context.Background(), fleet.HostListOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), authz.ForbiddenErrorMessage)
}

func TestGetHostSummary(t *testing.T) {
	ds := new(mock.Store)
	svc := newTestService(ds, nil, nil)

	ds.GenerateHostStatusStatisticsFunc = func(ctx context.Context, filter fleet.TeamFilter, now time.Time) (online uint, offline uint, mia uint, new uint, err error) {
		return 1, 2, 3, 4, nil
	}

	want := &fleet.HostSummary{OnlineCount: 1, OfflineCount: 2, MIACount: 3, NewCount: 4}
	summary, err := svc.GetHostSummary(test.UserContext(test.UserAdmin))
	require.NoError(t, err)
	require.Equal(t, summary, want)

	summary, err = svc.GetHostSummary(test.UserContext(test.UserNoRoles))
	require.NoError(t, err)
	require.Equal(t, summary, want)
}
