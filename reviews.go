package main

import (
	"fmt"
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

func (s *server) getAllReviews() {
	for {
		listSyncChannel := make(chan reviewsGetResponse)
		time.Sleep(time.Duration(s.config.GetListTime) * time.Second)
		for _, packageName := range s.packageList {
			go s.getReviews(packageName, listSyncChannel)
		}
		shouldSave := false
		for range s.packageList {
			getResponse := <-listSyncChannel
			if getResponse.list == nil {
				continue
			}
			s.control.reviewsMutex.Lock()
			defer s.control.reviewsMutex.Unlock()
			count, updates, new := mergeReviewLists(s.localReviews[getResponse.packageName], getResponse.list, getResponse.packageName)
			s.updateAlerts(getResponse.packageName, updates, new)
			shouldSave = shouldSave || count > 0
		}
		if shouldSave {
			s.SaveReviews()
		}
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
		fmt.Print(list.Reviews)
		listSyncChannel <- reviewsGetResponse{
			packageName: packageName,
			list:        list.Reviews,
		}
	}
}

func formatReview(review *androidpublisher.Review) string {
	stars := [...]string{
		"",
		":star:",
		":star::star:",
		":star::star::star:",
		":star::star::star::star:",
		":star::star::star::star::star:",
	}

	lastModified := review.Comments[0].UserComment.LastModified
	return fmt.Sprintf("%s commented (%s):\n%s\non %s\nReviewId:%s\n",
		review.AuthorName,
		stars[review.Comments[0].UserComment.StarRating],
		review.Comments[0].UserComment.Text,
		time.Unix(lastModified.Seconds, lastModified.Nanos),
		review.ReviewId)
}

func mergeReviewLists(localList []*androidpublisher.Review, remoteList []*androidpublisher.Review, packageName string) (int, []*androidpublisher.Review, []*androidpublisher.Review) {
	// Remove already cached elements
	for i := range remoteList {
		if remoteList[i].Comments[0].UserComment.LastModified.Seconds <= localList[0].Comments[0].UserComment.LastModified.Seconds {
			remoteList = remoteList[:i-1]
			break
		}
	}

	updatedReviews := []*androidpublisher.Review{}
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

	var newReviews []*androidpublisher.Review
	copy(newReviews, remoteList)

	// Join lists
	localList = append(remoteList, localList...)

	return 0, updatedReviews, newReviews
}
