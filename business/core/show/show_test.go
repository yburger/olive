package show_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-olive/olive/business/core/show"
	"github.com/go-olive/olive/business/data/dbtest"
	"github.com/go-olive/olive/foundation/docker"
	"github.com/google/go-cmp/cmp"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_Show(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testshow")
	t.Cleanup(teardown)

	core := show.NewCore(log, db)

	t.Log("Given the need to work with Show records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Show.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			nu := show.NewShow{
				Enable:    true,
				Platform:  "bilibili",
				RoomID:    "21852",
				PostCmds:  "[]",
				SplitRule: `{"FileSize": 1024}`,
			}

			s, err := core.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create show : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create show.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, s.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve show by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve show by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(s, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same show. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same show.", dbtest.Success, testID)

			upd := show.UpdateShow{
				Enable:    dbtest.BoolPointer(false),
				SplitRule: dbtest.StringPointer(`{"FileSize": 1024000}`),
			}

			if err := core.Update(ctx, s.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update show : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update show.", dbtest.Success, testID)

			saved, err = core.QueryByID(ctx, s.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve show by ID : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve show by Email.", dbtest.Success, testID)

			if saved.Enable != *upd.Enable {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Enable.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Enable)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Enable)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", dbtest.Success, testID)
			}

			if err := core.Delete(ctx, s.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete show : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete show.", dbtest.Success, testID)

			_, err = core.QueryByID(ctx, s.ID)
			if !errors.Is(err, show.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve show : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve show.", dbtest.Success, testID)
		}
	}
}

// func Test_PagingShow(t *testing.T) {
// 	log, db, teardown := dbtest.NewUnit(t, c, "testpaging")
// 	t.Cleanup(teardown)

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	dbschema.Seed(ctx, db)

// 	show := show.NewCore(log, db)

// 	t.Log("Given the need to page through Show records.")
// 	{
// 		testID := 0
// 		t.Logf("\tTest %d:\tWhen paging through 2 shows.", testID)
// 		{
// 			ctx := context.Background()

// 			shows1, err := show.Query(ctx, 1, 1)
// 			if err != nil {
// 				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve shows for page 1 : %s.", dbtest.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould be able to retrieve shows for page 1.", dbtest.Success, testID)

// 			if len(shows1) != 1 {
// 				t.Fatalf("\t%s\tTest %d:\tShould have a single show : %s.", dbtest.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould have a single show.", dbtest.Success, testID)

// 			shows2, err := show.Query(ctx, 2, 1)
// 			if err != nil {
// 				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve shows for page 2 : %s.", dbtest.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould be able to retrieve shows for page 2.", dbtest.Success, testID)

// 			if len(shows2) != 1 {
// 				t.Fatalf("\t%s\tTest %d:\tShould have a single show : %s.", dbtest.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould have a single show.", dbtest.Success, testID)

// 			if shows1[0].ID == shows2[0].ID {
// 				t.Logf("\t\tTest %d:\tShow1: %v", testID, shows1[0].ID)
// 				t.Logf("\t\tTest %d:\tShow2: %v", testID, shows2[0].ID)
// 				t.Fatalf("\t%s\tTest %d:\tShould have different shows : %s.", dbtest.Failed, testID, err)
// 			}
// 			t.Logf("\t%s\tTest %d:\tShould have different shows.", dbtest.Success, testID)
// 		}
// 	}
// }
