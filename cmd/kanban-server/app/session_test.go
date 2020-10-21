package app

import (
	"path/filepath"
)

var (
	aionDataPath, _  = filepath.Abs("./../../../test/test_data")
	microserviceName = "test"
)

// func Test_NormalCase001_GetKanban_WithFile(t *testing.T) {
// 	ctx := context.Background()
//
// 	s := NewMicroserviceSessionWithFile(ctx, aionDataPath)
// 	p := &kanbanpb.InitializeService{
// 		MicroserviceName: microserviceName,
// 		ProcessNumber:    1,
// 	}
// 	res := &kanbanpb.Response{}
//
// 	s.ReadKanban(p, res)
// 	if res.Error != "" {
// 		t.Fatal(res.Error)
// 	}
// 	// check exist file
// 	fileList, err := filepath.Glob(path.Join(aionDataPath, microserviceName, "C_*.json"))
// 	if err != nil || len(fileList) == 0 {
// 		t.Error(err)
// 	}
// 	filePath := fileList[0]
//
// 	if res.MessageType != kanbanpb.ResponseType_RES_CACHE_KANBAN {
// 		t.Errorf("cant set response type %s", kanbanpb.ResponseType_RES_CACHE_KANBAN)
// 	}
// 	respKanban := &kanbanpb.StatusKanban{}
// 	openKanban := &kanbanpb.StatusKanban{}
// 	if err := ptypes.UnmarshalAny(res.Message, openKanban); err != nil {
// 		t.Error(err)
// 	}
// 	r, err := os.Open(filePath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer r.Close()
// 	jsonpb.Unmarshal(r, respKanban)
//
// 	if respKanban.Services[0] != openKanban.Services[0] {
// 		t.Errorf("invalid open yaml ")
// 	}
// }
