package models

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gobuffalo/nulls"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v6"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"
}

func UploadFileToS3(video *Video, filePath string, s3Key string) error {
	//accessKeyID := os.Getenv("BUCKETEER_AWS_ACCESS_KEY_ID")
	//secretAccessKey := os.Getenv("BUCKETEER_AWS_SECRET_ACCESS_KEY")
	//bucketName := os.Getenv("SPACES_BUCKET")
	//awsRegion := os.Getenv("BUCKETEER_AWS_REGION")
	//
	//if accessKeyID == "" || secretAccessKey == "" || bucketName == "" || awsRegion == "" {
	//	return fmt.Errorf("Environment variables for S3 not properly set")
	//}
	//
	//// Create a new AWS session with the provided credentials and region
	//sess, err := session.NewSession(&aws.Config{
	//	Region:      aws.String(awsRegion),
	//	Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	//})

	sess, err := s3Session()
	bucketName := os.Getenv("SPACES_BUCKET")

	if err != nil {
		fmt.Printf("Failed to create AWS session %s", err)
		return fmt.Errorf("Failed to create AWS session: %v", err)
	}

	// Create a new S3 client using the AWS session
	svc := s3.New(sess)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Upload the file to S3

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		fmt.Printf("Failed to upload file to S3: %v\n", err)
		return err
	}

	fmt.Printf("Video '%s' uploaded to S3 with key '%s' in bucket '%s'.\n", filePath, s3Key, bucketName)

	uploadedAt := nulls.NewTime(time.Now())
	video.UploadedAt = uploadedAt

	err = DB.Update(video)
	if err != nil {
		fmt.Printf("Failed to update the video with the uploaded at timestamp %s\n", err)
	}

	return nil
}

func DownloadFileFromS3(objectKey string) (string, error) {
	//accessKeyID := os.Getenv("BUCKETEER_AWS_ACCESS_KEY_ID")
	//secretAccessKey := os.Getenv("BUCKETEER_AWS_SECRET_ACCESS_KEY")
	//bucketName := os.Getenv("SPACES_BUCKET")
	//awsRegion := os.Getenv("BUCKETEER_AWS_REGION")
	//
	//// Create a new AWS session with the provided credentials and region
	//sess, err := session.NewSession(&aws.Config{
	//	Region:      aws.String(awsRegion),
	//	Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	//})
	sess, err := s3Session()
	bucketName := os.Getenv("SPACES_BUCKET")

	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create a new S3 client using the AWS session
	svc := s3.New(sess)

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "downloaded_file_*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Prepare the S3 input parameters
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	// Download the file from S3 and save it to the temporary file
	resp, err := svc.GetObject(params)
	if err != nil {
		return "", fmt.Errorf("failed to download file from S3: %v", err)
	}
	defer resp.Body.Close()

	// Copy the content from S3 response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy S3 object to temporary file: %v", err)
	}

	fmt.Printf("Saved %s to %s\n", objectKey, tempFile.Name())
	return tempFile.Name(), nil
}

func uploadFolderToS3(folderPrefix string, localFolderPath string) error {
	sess, err := s3Session()
	bucketName := os.Getenv("SPACES_BUCKET")

	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create a new S3 client using the AWS session
	svc := s3.New(sess)

	// Walk through the local folder and upload each file to S3
	err = filepath.Walk(localFolderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, as we only want to upload files
		if info.IsDir() {
			return nil
		}

		// Determine the S3 key (object key) by removing the localFolderPath from the filePath
		// This ensures the relative path inside the folder is used as the key
		relPath, err := filepath.Rel(localFolderPath, filePath)
		if err != nil {
			return err
		}

		reader, _ := os.Open(filePath)

		// Prepare the S3 input parameters
		params := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(folderPrefix + relPath),
			Body:   aws.ReadSeekCloser(reader),
			ACL:    aws.String("public-read"),
		}

		// Upload the file to S3
		_, err = svc.PutObject(params)
		if err != nil {
			return fmt.Errorf("failed to upload file to S3: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to upload folder: %v", err)
	}

	return nil
}

func s3Session() (*session.Session, error) {
	key := os.Getenv("SPACES_KEY")
	secretAccessKey := os.Getenv("SPACES_SECRET")
	endpoint := os.Getenv("SPACES_ENDPOINT")

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(false), // // Configures to use subdomain/virtual calling format. Depending on your version, alternatively use o.UsePathStyle = false
	}

	fmt.Printf("Creating new s3 session...\n")

	sess, err := session.NewSession(s3Config)

	return sess, err
}

func VideoSchedule() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Create a channel to handle OS signals (e.g., SIGINT, SIGTERM)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt) // Subscribe to SIGINT (Ctrl+C) signals

	for {
		select {
		case <-ticker.C:
			err := ProcessVideos()
			if err != nil {
				fmt.Printf("Error occurred while processing videos %s\n", err)
			}
		case <-signalCh:
			fmt.Println("Received OS signal. Stopping...")
			return
		}
	}
}

func ProcessVideos() error {
	video := &Video{}
	query := DB.Where("processed = false AND uploaded_at IS NOT NULL")
	err := query.First(video)

	if err != nil {
		return fmt.Errorf("Could not find video that needed processed %s\n", err)
	}

	// Download from s3 to tmp file
	videoId := video.ID.String()

	fmt.Printf("Downloading %s from S3\n", videoId)
	file, err := DownloadFileFromS3(videoId)
	if err != nil {
		return fmt.Errorf("failed to download file from s3 %s\n", videoId)
	}
	// Create folder in tmp resources
	newFolderPath := "/tmp/" + videoId
	fmt.Printf("Converting %s to output folder %s\n", file, newFolderPath)
	err = ConvertVideo(file, newFolderPath)

	if err != nil {
		return err
	}

	fmt.Printf("Uploading folder to S3\n")
	err = uploadFolderToS3(videoId+"/", newFolderPath)
	if err != nil {
		return err
	}

	video.Processed = true
	err = DB.Update(video)
	if err != nil {
		return fmt.Errorf("failed to save video to database %s", err)
	}

	return nil
}

func ConvertVideo(inputFile string, outputDir string) error {
	// Create the output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Convert to three different resolutions (qualities)
	resolutions := []string{"480x270", "640x360", "1280x720", "1920x1080"}

	for _, resolution := range resolutions {
		outputPath := fmt.Sprintf("%s/quality_%s.m3u8", outputDir, resolution)

		//ffmpeg -i /var/folders/12/f70bt57x0rx34jk7l3r526_00000gn/T/downloaded_file_1356103922.tmp -c:v libx264 -c:a aac -strict -2 -f hls -hls_time 10 -hls_list_size 0 /var/folders/12/f70bt57x0rx34jk7l3r526_00000gn/T//cac586d2-653a-4501-8601-6170cd0d018a/480x270.m3u8

		cmd := exec.Command("ffmpeg",
			"-i", inputFile,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-strict", "-2",
			"-f", "hls",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-vf", fmt.Sprintf("scale=%s", resolution),
			outputPath,
		)

		fmt.Printf("Running command %s \n", cmd.String())

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to convert video: %v", err)
		} else {
			fmt.Printf("Converted %s to %s\n", inputFile, outputPath)
		}
	}

	return nil
}
