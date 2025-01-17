// Copyright (C) 2022 Parseable, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"collector/pkg/client"
	"collector/pkg/collector"
	"collector/pkg/parseable"
	"encoding/json"

	"time"

	"os"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func RunKubeCollector(url, user, pwd string, stream *LogStream) {
	if err := parseable.CreateStream(url, user, pwd, stream.Name); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Infof("Successfully created Log Stream [%s] on server [%s]", stream.Name, url)
	interval, err := time.ParseDuration(stream.CollectInterval)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	ticker := time.NewTicker(interval)
	for range ticker.C {
		kubeCollector(url, user, pwd, stream)
	}
}

func kubeCollector(url, user, pwd string, stream *LogStream) {
	collectFrom := stream.CollectFrom
	var podsList []*v1.PodList
	for k, v := range collectFrom.PodSelector {
		pods, err := client.KubeClient.ListPods(collectFrom.Namespace, k+"="+v)
		if err != nil {
			log.Error(err)
			return
		}
		podsList = append(podsList, pods)
	}
	for _, po := range podsList {
		for _, p := range po.Items {
			logs, err := collector.GetPodLogs(p, url, user, pwd, stream.Name)
			if err != nil {
				log.Error(err)
				return
			}
			if len(logs) > 0 {
				if err != nil {
					log.Error(err)
					return
				} else {
					log.Infof("Successfully collected log from [%s] in [%s] namespace", p.GetName(), p.Namespace)
				}
				jLogs, err := json.Marshal(logs)
				if err != nil {
					return
				}
				if err := parseable.PostLogs(url, user, pwd, stream.Name, jLogs, stream.Labels); err != nil {
					log.Error(err)
				}
				log.Infof("Successfully sent log from [%s] in [%s] namespace to server [%s]", p.GetName(), p.GetNamespace(), url)
			}
		}
	}
}
