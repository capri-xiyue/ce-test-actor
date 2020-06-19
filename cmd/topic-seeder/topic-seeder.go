package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"
	"github.com/google/knative-gcp/pkg/utils"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

type config struct {
	TopicID     string        `envconfig:"TOPIC_ID"`
	Interval    time.Duration `envconfig:"INTERVAL"`
	Concurrency int           `envconfig:"CONCURRENCY" default:"1"`
}

func main() {
	var env config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env config: %v", err)
	}
	ctx := context.Background()
	projectID, err := utils.ProjectID(os.Getenv("PROJECT_ID"), metadataClient.NewDefaultMetadataClient())
	if err != nil {
		log.Fatalf("Failed to get project ID", err)
	}
	log.Infof("projectID is %q \n", projectID)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to get pubsub client", err)
	}
	t := client.Topic(env.TopicID)
	log.Infof("topicID is %q \n", env.TopicID)
	for {
		for i := 0; i < env.Concurrency; i++ {
			go func() {
				result := t.Publish(ctx, &pubsub.Message{
					Data: []byte("Message " + strconv.Itoa(i)),
				})
				id, err := result.Get(ctx)
				if err != nil {
					log.Errorf("Failed to publish message (id=%q) to topic %q: %v", i, env.TopicID, err)
				}
				log.Infof("Published message with custom attributes; msg ID: %v\n", id)

			}()
		}

		log.Infof("Sleeping...")
		time.Sleep(env.Interval)
	}
}

