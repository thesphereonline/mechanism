package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

var (
	s3Client  *s3.S3
	ipfsShell *ipfsapi.Shell
	s3Bucket  string
	useIPFS   bool
)

// InitStorage initializes the storage service
func InitStorage() error {
	// Check if we should use IPFS
	useIPFSStr := os.Getenv("USE_IPFS")
	if useIPFSStr == "true" {
		useIPFS = true
		ipfsURL := os.Getenv("IPFS_URL")
		if ipfsURL == "" {
			ipfsURL = "http://localhost:5001"
		}
		ipfsShell = ipfsapi.NewShell(ipfsURL)
	} else {
		// Initialize S3 client
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}

		s3Bucket = os.Getenv("S3_BUCKET")
		if s3Bucket == "" {
			return fmt.Errorf("S3 bucket name not set")
		}

		awsAccessKey := os.Getenv("AWS_ACCESS_KEY")
		awsSecretKey := os.Getenv("AWS_SECRET_KEY")

		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS session: %v", err)
		}

		s3Client = s3.New(sess)
	}

	return nil
}

// UploadFile uploads a file to storage and returns the URL
func UploadFile(file multipart.File, filename string) (string, error) {
	if useIPFS {
		return uploadToIPFS(file)
	}
	return uploadToS3(file, filename)
}

// UploadMetadata uploads NFT metadata to storage and returns the URL
func UploadMetadata(metadata map[string]interface{}) (string, error) {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if useIPFS {
		return uploadJSONToIPFS(metadataJSON)
	}
	return uploadJSONToS3(metadataJSON)
}

// uploadToS3 uploads a file to S3
func uploadToS3(file multipart.File, filename string) (string, error) {
	// Reset file pointer to beginning
	if seeker, ok := file.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", fmt.Errorf("failed to reset file pointer: %v", err)
		}
	}

	// Get file content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	contentType := http.DetectContentType(buffer)

	// Reset file pointer again
	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", fmt.Errorf("failed to reset file pointer: %v", err)
		}
	}

	// Upload to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// Generate URL
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Bucket, filename)
	return s3URL, nil
}

// uploadJSONToS3 uploads JSON data to S3
func uploadJSONToS3(data []byte) (string, error) {
	// Generate unique filename
	filename := fmt.Sprintf("metadata/%s.json", uuid.New().String())

	// Upload to S3
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload metadata to S3: %v", err)
	}

	// Generate URL
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Bucket, filename)
	return s3URL, nil
}

// uploadToIPFS uploads a file to IPFS
func uploadToIPFS(file multipart.File) (string, error) {
	// Reset file pointer to beginning
	if seeker, ok := file.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", fmt.Errorf("failed to reset file pointer: %v", err)
		}
	}

	// Upload to IPFS
	cid, err := ipfsShell.Add(file)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to IPFS: %v", err)
	}

	// Generate URL
	ipfsURL := fmt.Sprintf("ipfs://%s", cid)
	return ipfsURL, nil
}

// uploadJSONToIPFS uploads JSON data to IPFS
func uploadJSONToIPFS(data []byte) (string, error) {
	// Upload to IPFS
	cid, err := ipfsShell.Add(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to upload metadata to IPFS: %v", err)
	}

	// Generate URL
	ipfsURL := fmt.Sprintf("ipfs://%s", cid)
	return ipfsURL, nil
}
