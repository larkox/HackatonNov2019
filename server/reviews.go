package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

type reviewsGetResponse = struct {
	packageName string
	userID      string
	list        []*androidpublisher.Review
}

var mockReview = androidpublisher.Review{
	AuthorName: "AuthorName",
	Comments: []*androidpublisher.Comment{
		&androidpublisher.Comment{
			UserComment: &androidpublisher.UserComment{
				AndroidOsVersion: 23,
				AppVersionCode:   1,
				AppVersionName:   "VersionName",
				Device:           "Device",
				DeviceMetadata:   nil,
				LastModified: &androidpublisher.Timestamp{
					Nanos:   1000000,
					Seconds: 1,
				},
				OriginalText:     "OriginalText",
				ReviewerLanguage: "ES",
				StarRating:       5,
				Text:             "Text",
				ThumbsDownCount:  1,
				ThumbsUpCount:    1,
			},
		},
	},
	ReviewId: "ReviewId",
}

func newMockReview() androidpublisher.Review {
	rand.Seed(time.Now().UnixNano())
	return androidpublisher.Review{
		AuthorName: "MockUser",
		Comments: []*androidpublisher.Comment{
			&androidpublisher.Comment{
				UserComment: &androidpublisher.UserComment{
					AndroidOsVersion: 23,
					AppVersionCode:   1,
					AppVersionName:   "VersionName",
					Device:           "Device",
					DeviceMetadata:   nil,
					LastModified: &androidpublisher.Timestamp{
						Nanos:   0,
						Seconds: time.Now().Unix(),
					},
					OriginalText:     "OriginalText",
					ReviewerLanguage: "ES",
					StarRating:       rand.Int63n(6),
					Text:             fmt.Sprintf("This is a mock review written at %v", time.Now()),
					ThumbsDownCount:  1,
					ThumbsUpCount:    1,
				},
			},
		},
		ReviewId: fmt.Sprintf("%d", time.Now().Unix()),
	}
}

func (p *Plugin) getAllReviews() {
	for {
		listSyncChannel := make(chan reviewsGetResponse)
		for _, packageInfo := range p.packageList {
			go p.getReviews(packageInfo.Name, packageInfo.UserID, listSyncChannel)
		}
		shouldSave := false
		for range p.packageList {
			getResponse := <-listSyncChannel
			if getResponse.list == nil || len(getResponse.list) == 0 {
				// mockReview := newMockReview()
				// getResponse.list = []*androidpublisher.Review{
				// 	&mockReview,
				// }
				continue
			}

			p.control.reviewsMutex.Lock()
			if _, ok := p.localReviews[getResponse.userID]; !ok {
				p.localReviews[getResponse.userID] = make(map[string][]*androidpublisher.Review)
			}
			if _, ok := p.localReviews[getResponse.userID][getResponse.packageName]; !ok {
				p.localReviews[getResponse.userID][getResponse.packageName] = []*androidpublisher.Review{}
			}
			count, local, updates, new := mergeReviewLists(p.localReviews[getResponse.userID][getResponse.packageName], getResponse.list, getResponse.packageName)
			p.localReviews[getResponse.userID][getResponse.packageName] = local
			p.updateAlerts(getResponse.packageName, getResponse.userID, updates, new)
			p.control.reviewsMutex.Unlock()

			shouldSave = shouldSave || count > 0
		}

		if shouldSave {
			p.control.reviewsMutex.Lock()
			p.SaveReviews()
			p.control.reviewsMutex.Unlock()
		}
		config := p.getConfiguration()
		time.Sleep(time.Duration(config.GetListTime) * time.Second)
	}
}

func (p *Plugin) getReviews(packageName string, userID string, listSyncChannel chan reviewsGetResponse) {
	service := p.getService(userID)
	list, err := service.List(packageName).Do()
	if err != nil {
		fmt.Print(err.Error())
		listSyncChannel <- reviewsGetResponse{
			packageName: packageName,
			userID:      userID,
			list:        nil,
		}
	} else {
		listSyncChannel <- reviewsGetResponse{
			packageName: packageName,
			userID:      userID,
			list:        list.Reviews,
		}
	}
}

func formatReview(review *androidpublisher.Review) string {
	stars := [...]string{
		":new_moon::new_moon::new_moon::new_moon::new_moon:",
		":star::new_moon::new_moon::new_moon::new_moon:",
		":star::star::new_moon::new_moon::new_moon:",
		":star::star::star::new_moon::new_moon:",
		":star::star::star::star::new_moon:",
		":star::star::star::star::star:",
	}

	lastModified := review.Comments[0].UserComment.LastModified
	return fmt.Sprintf("#### **%s** commented (%s):\n>%s\n\non _%s_\nReviewId:**%s**\n",
		review.AuthorName,
		stars[review.Comments[0].UserComment.StarRating],
		strings.Join(strings.Split(review.Comments[0].UserComment.Text, "\n"), "\n>"),
		time.Unix(lastModified.Seconds, lastModified.Nanos),
		review.ReviewId)
}

func mergeReviewLists(localList []*androidpublisher.Review, remoteList []*androidpublisher.Review, packageName string) (count int, newLocalList []*androidpublisher.Review, updatedReviews []*androidpublisher.Review, newReviews []*androidpublisher.Review) {
	// Remove already cached elements
	if len(localList) != 0 {
		for i := range remoteList {
			if remoteList[i].Comments[0].UserComment.LastModified.Seconds <= localList[0].Comments[0].UserComment.LastModified.Seconds {
				remoteList = remoteList[:i-1]
				break
			}
		}
	}

	updatedReviews = []*androidpublisher.Review{}
	// Remove duplicates
	for _, listItem := range remoteList {
		for i, cacheItem := range localList {
			if listItem.ReviewId == cacheItem.ReviewId {
				updatedReviews = append(updatedReviews, listItem)
				localList = removeElement(localList, i)
				break
			}
		}
	}

	newReviews = append([]*androidpublisher.Review(nil), remoteList...)

	// Join lists
	localList = append(remoteList, localList...)

	return 0, localList, updatedReviews, newReviews
}
