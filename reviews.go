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

func (s *server) getAllReviews() {
	for {
		listSyncChannel := make(chan reviewsGetResponse)
		for _, packageName := range s.packageList {
			go s.getReviews(packageName, listSyncChannel)
		}
		shouldSave := false
		for range s.packageList {
			getResponse := <-listSyncChannel
			if getResponse.list == nil || len(getResponse.list) == 0 {
				if s.config.useMock {
					mockReview := newMockReview()
					getResponse.list = []*androidpublisher.Review{
						&mockReview,
					}
				} else {
					continue
				}
			}

			s.control.reviewsMutex.Lock()
			count, local, updates, new := mergeReviewLists(s.localReviews[getResponse.packageName], getResponse.list, getResponse.packageName)
			s.localReviews[getResponse.packageName] = local
			s.updateAlerts(getResponse.packageName, updates, new)
			s.control.reviewsMutex.Unlock()

			shouldSave = shouldSave || count > 0
		}

		if shouldSave {
			s.control.reviewsMutex.Lock()
			s.SaveReviews()
			s.control.reviewsMutex.Unlock()
		}
		time.Sleep(time.Duration(s.config.GetListTime) * time.Second)
	}
}

func (s *server) getReviews(packageName string, listSyncChannel chan reviewsGetResponse) {
	list, err := s.service.Reviews.List(packageName).Do()
	if err != nil {
		fmt.Print(err.Error())
		listSyncChannel <- reviewsGetResponse{
			packageName: packageName,
			list:        nil,
		}
	} else {
		listSyncChannel <- reviewsGetResponse{
			packageName: packageName,
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
