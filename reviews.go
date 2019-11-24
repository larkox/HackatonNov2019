package main

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

// Newer reviews will always be on the lower ids of the slice
var localReviews = make(map[string][]*androidpublisher.Review)
var reviewsMutex sync.Mutex
var newReviewsCounts = make(map[string]int)

var stars = [...]string{
	"",
	":star:",
	":star::star:",
	":star::star::star:",
	":star::star::star::star:",
	":star::star::star::star::star:",
}

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

func getAllReviews() {
	for {
		//testAlert(&mockReview)
		listSyncChannel := make(chan reviewsGetResponse)
		time.Sleep(getListTime)
		for _, packageName := range packageList {
			go getReviews(packageName, listSyncChannel)
		}
		for range packageList {
			getResponse := <-listSyncChannel
			if getResponse.list == nil {
				continue
			}
			mergeReviewLists(localReviews[getResponse.packageName], getResponse.list, getResponse.packageName)
		}
		for k, v := range newReviewsCounts {
			if v > 0 {
				go alertNewReviews(k)
			}
		}
	}
}

func getReviews(packageName string, listSyncChannel chan reviewsGetResponse) {
	list, err := service.Reviews.List(packageName).Do()
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
	lastModified := review.Comments[0].UserComment.LastModified
	return fmt.Sprintf("%s commented (%s):\\n%s\\non %s\\nReviewId:%s\\n",
		review.AuthorName,
		stars[review.Comments[0].UserComment.StarRating],
		review.Comments[0].UserComment.Text,
		time.Unix(lastModified.Seconds, lastModified.Nanos),
		review.ReviewId)
}

func mergeReviewLists(localList []*androidpublisher.Review, remoteList []*androidpublisher.Review, packageName string) {
	reviewsMutex.Lock()
	// Remove already cached elements
	for i := range remoteList {
		if remoteList[i].Comments[0].UserComment.LastModified.Seconds <= localList[0].Comments[0].UserComment.LastModified.Seconds {
			remoteList = remoteList[:i-1]
			break
		}
	}

	// Remove duplicates
	for _, listItem := range remoteList {
		for i, cacheItem := range localList {
			if listItem.ReviewId == cacheItem.ReviewId {
				localList = removeElement(localList, i)
				break
			}
		}
	}

	newReviewsCounts[packageName] += len(remoteList)

	// Join lists
	localList = append(remoteList, localList...)
	reviewsMutex.Unlock()
}
