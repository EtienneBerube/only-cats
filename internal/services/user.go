package services

import (
	"errors"
	"fmt"
	"github.com/EtienneBerube/cat-scribers/internal/models"
	"github.com/EtienneBerube/cat-scribers/internal/repositories"
	"log"
)

func GetUserById(id string) (*models.User, error) {
	user, err := repositories.GetUserById(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetAllUsers() ([]models.User, error) {
	users, err := repositories.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func PaySubscriptionTo(user *models.User, subscribedToID string) error {
	subscribedToUser, err := repositories.GetUserById(subscribedToID)
	if err != nil {
		UnsubscribeFrom(user.ID, subscribedToUser.ID)
		log.Printf("ERROR ON CRON: %s", err.Error())
	}

	if user.Balance-subscribedToUser.SubscriptionPrice < 0 {
		UnsubscribeFrom(user.ID, subscribedToUser.ID)
		log.Printf("CRON: %s cannot pay for %s's subscription. Unsubscribing...", user.Name, subscribedToUser.Name)

	} else {
		user.Balance -= subscribedToUser.SubscriptionPrice
		subscribedToUser.Balance += subscribedToUser.SubscriptionPrice
		ok, err := repositories.UpdateUser(subscribedToUser.ID, subscribedToUser)
		if err != nil || !ok {
			log.Printf("CRON: Error while giving the money to %s's. Error: %s", subscribedToUser.Name, err.Error())
		}
	}
	ok, err := repositories.UpdateUser(user.ID, user)
	if err != nil || !ok {
		log.Printf("CRON: Error while updating %s's balance. Error: %s", user.Name, err.Error())
	}
	return err
}

func PaySubscription(user *models.User) {
	for _, subscribedToID := range user.Subscriptions {
		err := PaySubscriptionTo(user, subscribedToID)
		if err != nil {
			log.Printf("ERROR ON CRON: %s", err.Error())
			continue
		}
	}
}


func CreateNewUser(user models.User) (string, error) {
	id, err := repositories.SaveUser(user)
	if err != nil {
		return "", err
	}
	return id, nil
}

func UpdateUser(currentUserID string, user models.User) (bool, error) {
	currentUser, err := GetUserById(currentUserID)
	if err != nil {
		return false, err
	}
	if currentUser.Email != user.Email {
		return false, errors.New("Cannot change user email as it is used for auth")
	}

	return repositories.UpdateUser(currentUserID, &user)
}

func SubscribeTo(currentUserID string, newSubscriptionID string) (bool, error) {
	currentUser, err := GetUserById(currentUserID)
	if err != nil {
		return false, err
	}

	currentUser.Subscriptions = append(currentUser.Subscriptions, newSubscriptionID)

	ok, err := repositories.UpdateUser(currentUserID, currentUser)
	if err != nil {
		return false, err
	}

	PaySubscriptionTo(currentUser, newSubscriptionID)

	return ok, nil
}

func UnsubscribeFrom(currentUserID string, subscriptionIDToRemove string) (bool, error) {
	currentUser, err := GetUserById(currentUserID)
	if err != nil {
		return false, err
	}
	var removeIndex int
	for index, id := range currentUser.Subscriptions {
		if id == subscriptionIDToRemove {
			removeIndex = index
			break
		}
	}

	currentUser.Subscriptions = append(currentUser.Subscriptions[:removeIndex], currentUser.Subscriptions[removeIndex+1:]...)

	ok, err := repositories.UpdateUser(currentUserID, currentUser)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func DeleteUser(id string) error {
	subscribers, err := repositories.GetAllUsersSubscribedTo(id)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot get all users subscribed to %s", id))
	}
	for _, subscriber := range subscribers {
		ok, err := UnsubscribeFrom(subscriber.ID, id)
		if err != nil || !ok {
			return errors.New(fmt.Sprintf("Cannot unsubscribe others from id:%s account. Error: %s", id, err.Error()))
		}
	}

	return repositories.DeleteUser(id)
}
