package paraformer

import (
	"context"
	"framework/logger"
	"net/http"
	"strings"

	httpclient "github.com/devinyf/dashscopego/httpclient"
	"github.com/google/uuid"
)

// real-time voice recognition

func ConnRecognitionClient(request *Request, token string) (*httpclient.WsClient, error) {
	// Initialize the client with the necessary parameters.
	header := http.Header{}
	header.Add("Authorization", token)

	client := httpclient.NewWsClient(ParaformerWSURL, header)

	if err := client.ConnClient(request); err != nil {
		return nil, err
	}

	return client, nil
}

func CloseRecognitionClient(cli *httpclient.WsClient) error {
	if err := cli.CloseClient(); err != nil {
		logger.Debugf("close client error: %v", err)
		return err
	}

	return nil
}

func SendRadioData(cli *httpclient.WsClient, bytesData []byte) {
	cli.SendBinaryDates(bytesData)
}

type ResultWriter interface {
	WriteResult(str string) error
}

func HandleRecognitionResult(ctx context.Context, cli *httpclient.WsClient, fn StreamingFunc) {
	outputChan, errChan := cli.ResultChans()

	// TODO: handle errors.
BREAK_FOR:
	for {
		select {
		case output, ok := <-outputChan:
			if !ok {
				logger.Debugf("outputChan is closed")
				break BREAK_FOR
			}

			// streaming callback func
			if err := fn(ctx, output.Data); err != nil {
				logger.Errorf("error: ", err)
				break BREAK_FOR
			}

		case err := <-errChan:
			if err != nil {
				logger.Errorf("error: ", err)
				break BREAK_FOR
			}
		case <-ctx.Done():
			cli.Over = true
			logger.Debugf("Done")
			break BREAK_FOR
		}
	}

	logger.Debugf("get recognition result...over")
}

// task_id length 32.
func GenerateTaskID() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	uuid := strings.ReplaceAll(u.String(), "-", "")

	return uuid
}
