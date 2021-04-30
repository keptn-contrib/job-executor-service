package k8s



/*
	client.pods().inNamespace(namespace).withName(podList.getItems().get(0).getMetadata().getName())
	            .waitUntilCondition(pod -> pod.getStatus().getPhase().equals("Succeeded"), 1, TimeUnit.MINUTES);

	          // Print Job's log
	          String joblog = client.batch().v1().jobs().inNamespace(namespace).withName("pi").getLog();
*/
