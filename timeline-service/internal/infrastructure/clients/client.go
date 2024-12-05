package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"
)

type client struct {
	userUrl  string
	tweetUrl string
}

func NewClient() interfaces.Client {
	return &client{
		userUrl:  "http://localhost:8080/users",
		tweetUrl: "http://localhost:8081/tweets",
	}
}

func (api *client) Tweets(ctx context.Context, page, limit int) ([]*models.Tweet, error) {
	url := fmt.Sprintf("%s?page=%d&limit=%d", api.tweetUrl, page, limit)

	fmt.Println(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Agrega el header personalizado
	req.Header.Set("User-ID", "2a42c7ae-7f78-4e36-8358-902342fe23f1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tweets []*models.Tweet
	if err := json.NewDecoder(resp.Body).Decode(&tweets); err != nil {
		return nil, err
	}
	return tweets, nil
}

func (api *client) User(ctx context.Context, userID string) (*models.User, error) {
	url := fmt.Sprintf("%s/%s", api.userUrl, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Agrega el header personalizado
	req.Header.Set("User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (api *client) Followers(ctx context.Context, userID string, page, limit int) ([]*models.User, error) {
	url := fmt.Sprintf("%s/%s/followers?page=%d&limit=%d", api.userUrl, userID, page, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Agrega el header personalizado
	req.Header.Set("User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var followers []*models.User
	if err := json.NewDecoder(resp.Body).Decode(&followers); err != nil {
		return nil, err
	}
	return followers, nil
}
