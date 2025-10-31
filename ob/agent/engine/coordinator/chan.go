/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package coordinator

type coordinatorEventChan struct {
	channel chan bool
	Close   func()
}

func (cChan *coordinatorEventChan) Listen() <-chan bool {
	return cChan.channel
}

func (c *Coordinator) Subscribe(obj interface{}) *coordinatorEventChan {
	if eventChan, ok := c.eventChans[obj]; ok {
		return eventChan
	}

	eventChan := &coordinatorEventChan{
		channel: make(chan bool, 1),
		Close: func() {
			c.unsubscribe(obj)
		},
	}
	c.eventChans[obj] = eventChan
	return eventChan
}

func (c *Coordinator) unsubscribe(obj interface{}) {
	if eventChan, ok := c.eventChans[obj]; ok {
		delete(c.eventChans, obj)
		eventChan.Close()
	}
}

func (c *Coordinator) publish(event bool) {
	for _, v := range c.eventChans {
		go func(v *coordinatorEventChan) { v.channel <- event }(v)
	}
}
