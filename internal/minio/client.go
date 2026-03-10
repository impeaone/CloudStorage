package minio_client

import (
	"CloudStorageProject-FileServer/internal/metrics"
	MinioConfig "CloudStorageProject-FileServer/internal/minio/config"
	"CloudStorageProject-FileServer/pkg/config"
	"CloudStorageProject-FileServer/pkg/models"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	MinioClient *minio.Client
	MinioConfig *MinioConfig.MinioConfig
	Metrics     *metrics.MinIOMetrics
	ctx         context.Context
}

func NewMinioClient(ctx context.Context, metric *metrics.MinIOMetrics) *MinioClient {
	conf := ctx.Value("config").(*config.Config)
	return &MinioClient{
		MinioConfig: MinioConfig.LoadMinioConfig(conf),
		Metrics:     metric,
		ctx:         ctx,
	}
}

func (mc *MinioClient) CloseConnection(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		mc.MinioClient.TraceOff()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (mc *MinioClient) Init() error {

	client, err := minio.New(mc.MinioConfig.MinioEndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(mc.MinioConfig.MinioRootUser, mc.MinioConfig.MinioRootPassword, ""),
		Secure: mc.MinioConfig.MinioUserSSL,
	})
	if err != nil {
		return err
	}
	mc.MinioClient = client

	exists, errTest := mc.MinioClient.BucketExists(mc.ctx, mc.MinioConfig.MinioExampleBucket)
	if errTest != nil {
		return errTest
	}

	if !exists {
		errNewTestBucket := mc.MinioClient.MakeBucket(mc.ctx, mc.MinioConfig.MinioExampleBucket, minio.MakeBucketOptions{})
		if errNewTestBucket != nil {
			return errNewTestBucket
		}
	}
	return nil
}

func (mc *MinioClient) CreateOne(apiBucket string, file models.FileMinio) error {

	start := time.Now()
	_, err := mc.MinioClient.PutObject(mc.ctx, apiBucket, file.FileName, file.Reader, file.Size, minio.PutObjectOptions{
		ContentType: file.ContentType,
	})
	end := time.Since(start)
	if err != nil {
		// если ошибка, добавляем метрики ошибок
		mc.Metrics.UploadErrors.WithLabelValues(apiBucket, err.Error()).Inc()
		return err
	}
	// Если добавление файла успешно, обновляем метрики
	// +1 к общему количеству загрузок
	mc.Metrics.UploadsTotal.WithLabelValues(apiBucket).Inc()
	// Запоминаем время загрузки
	mc.Metrics.UploadTime.WithLabelValues(apiBucket).Observe(end.Seconds())
	// Запоминаем размер файла
	mc.Metrics.UploadSize.WithLabelValues(apiBucket).Observe(float64(file.Size))
	return nil
}

// GetOne - берет файл с minio, взовращаем object потому что потом сразу в io.Writer, http.ResponseWriter
func (mc *MinioClient) GetOne(apiBucket string, objectName string) (*minio.Object, error) {

	obj, err := mc.MinioClient.GetObject(mc.ctx, apiBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		// Если ошибка выдачи файла
		mc.Metrics.DownloadErrors.WithLabelValues(apiBucket, err.Error()).Inc()
		return nil, err
	}
	// Если успешно, +1 к скачиваниям
	mc.Metrics.DownloadsTotal.WithLabelValues(apiBucket).Inc()
	return obj, nil
}

func (mc *MinioClient) FilesList(apiBucket string) ([]models.FileWebResponse, error) {

	var files []models.FileWebResponse

	objs := mc.MinioClient.ListObjects(mc.ctx, apiBucket, minio.ListObjectsOptions{
		Recursive: false,
	})
	for obj := range objs {
		if obj.Err != nil {
			mc.Metrics.FilesListErrors.WithLabelValues(apiBucket, obj.Err.Error()).Inc()
			return []models.FileWebResponse{}, obj.Err
		}
		fileSize := tools.FormatFileSize(obj.Size)
		fileName := obj.Key
		fileNameSplit := strings.Split(fileName, ".")
		fileType := fileNameSplit[len(fileNameSplit)-1]
		files = append(files, models.FileWebResponse{
			FileName:    fileName,
			FileSize:    fileSize,
			FileType:    fileType,
			LastModTime: obj.LastModified.Format("02.01.2006 12:05"),
		})
	}
	mc.Metrics.FilesListTotal.WithLabelValues(apiBucket).Add(float64(len(files)))
	return files, nil
}

func (mc *MinioClient) Delete(apiBucket string, objectName string) error {

	err := mc.MinioClient.RemoveObject(mc.ctx, apiBucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		mc.Metrics.DeleteErrors.WithLabelValues(apiBucket, err.Error()).Inc()
		return err
	}
	mc.Metrics.DeletesTotal.WithLabelValues(apiBucket).Inc()
	return nil
}
