// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"reflect"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/api"
)

// processNotifSubscribeRequest processes a notification subscribe request.
func (c *Connection) processNotifSubscribeRequest(ch *api.Channel, req *api.NotifSubscribeRequest) error {
	var err error

	// subscribe / unsubscribe
	if req.Subscribe {
		err = c.addNotifSubscription(req.Subscription)
	} else {
		err = c.removeNotifSubscription(req.Subscription)
	}

	// send the reply into the go channel
	select {
	case ch.NotifSubsReplyChan <- err:
		// reply sent successfully
	default:
		// unable to write into the channel without blocking
		log.WithFields(logger.Fields{
			"channel": ch,
		}).Warn("Unable to deliver the subscribe reply, reciever end not ready.")
	}

	return err
}

// addNotifSubscription adds the notification subscription into the subscriptions map of the connection.
func (c *Connection) addNotifSubscription(subs *api.NotifSubscription) error {
	// get message ID of the notification message
	msgID, err := c.getSubscriptionMessageID(subs)
	if err != nil {
		return err
	}

	log.WithFields(logger.Fields{
		"msg_id":       msgID,
		"subscription": subs,
	}).Debug("Adding new notification subscription.")

	// add the subscription into map
	c.notifSubscriptionsLock.Lock()
	defer c.notifSubscriptionsLock.Unlock()

	c.notifSubscriptions[msgID] = append(c.notifSubscriptions[msgID], subs)

	return nil
}

// removeNotifSubscription removes the notification subscription from the subscriptions map of the connection.
func (c *Connection) removeNotifSubscription(subs *api.NotifSubscription) error {
	// get message ID of the notification message
	msgID, err := c.getSubscriptionMessageID(subs)
	if err != nil {
		return err
	}

	log.WithFields(logger.Fields{
		"msg_id":       msgID,
		"subscription": subs,
	}).Debug("Removing notification subscription.")

	// remove the subscription from the map
	c.notifSubscriptionsLock.Lock()
	defer c.notifSubscriptionsLock.Unlock()

	for i, item := range c.notifSubscriptions[msgID] {
		if item == subs {
			// remove i-th item in the slice
			c.notifSubscriptions[msgID] = append(c.notifSubscriptions[msgID][:i], c.notifSubscriptions[msgID][i+1:]...)
			break
		}
	}

	return nil
}

// isNotificationMessage returns true if someone has subscribed to provided message ID.
func (c *Connection) isNotificationMessage(msgID uint16) bool {
	c.notifSubscriptionsLock.RLock()
	defer c.notifSubscriptionsLock.RUnlock()

	_, exists := c.notifSubscriptions[msgID]
	return exists
}

// sendNotifications send a notification message to all subscribers subscribed for that message.
func (c *Connection) sendNotifications(msgID uint16, data []byte) {
	c.notifSubscriptionsLock.RLock()
	defer c.notifSubscriptionsLock.RUnlock()

	matched := false

	// send to notification to each subscriber
	for _, subs := range c.notifSubscriptions[msgID] {
		log.WithFields(logger.Fields{
			"msg_id":       msgID,
			"msg_size":     len(data),
			"subscription": subs,
		}).Debug("Sending a notification to the subscription channel.")

		msg := subs.MsgFactory()
		err := c.codec.DecodeMsg(data, msg)
		if err != nil {
			log.WithFields(logger.Fields{
				"msg_id":       msgID,
				"msg_size":     len(data),
				"subscription": subs,
			}).Error("Unable to decode the notification message.")
			continue
		}

		// special case for the strange interface counters message
		if msg.GetMessageName() == "vnet_interface_counters" {
			v := reflect.ValueOf(msg).Elem().FieldByName("Data")
			if v.IsValid() {
				v.SetBytes(data[8:]) // include the Count and Data fields in the data
			}
		}

		// send the message into the go channel of the subscription
		select {
		case subs.NotifChan <- msg:
			// message sent successfully
		default:
			// unable to write into the channel without blocking
			log.WithFields(logger.Fields{
				"msg_id":       msgID,
				"msg_size":     len(data),
				"subscription": subs,
			}).Warn("Unable to deliver the notification, reciever end not ready.")
		}

		matched = true
	}

	if !matched {
		log.WithFields(logger.Fields{
			"msg_id":   msgID,
			"msg_size": len(data),
		}).Debug("No subscription found for the notification message.")
	}
}

// getSubscriptionMessageID returns ID of the message the subscription is tied to.
func (c *Connection) getSubscriptionMessageID(subs *api.NotifSubscription) (uint16, error) {
	msg := subs.MsgFactory()
	msgID, err := c.GetMessageID(msg)

	if err != nil {
		log.WithFields(logger.Fields{
			"msg_name": msg.GetMessageName(),
			"msg_crc":  msg.GetCrcString(),
		}).Errorf("unable to retrieve message ID: %v", err)
		return 0, fmt.Errorf("unable to retrieve message ID: %v", err)
	}

	return msgID, nil
}
