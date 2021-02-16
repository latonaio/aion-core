package uploader

import (
	"bitbucket.org/latonaio/aion-core/proto/projectpb"
	"github.com/pkg/errors"
	"io"
)

type Uploader struct {
}

func (u *Uploader) Upload(stream projectpb.YamlUpload_UploadServer) error {
	file := []byte{}
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				goto END
			}

			return errors.Wrap(err, "failed unexpectadely while reading chunk from stream")
		}

		file = append(file, req.GetContent()...)
	}
END:
	err := stream.SendAndClose(&projectpb.UploadStatus{
		Message: "Success",
		Code:    projectpb.UploadStatusCode_OK,
	})
	if err != nil {
		return err
	}
	return nil
}
