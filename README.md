# ARC cleaner

ARC cleaner is an application to clean up resources from the GitHub
Actions Runner Controller (ARC).

GitHub Actions Runners in kubernetes mode sometimes get stuck as ephemeral
volumes are used. These are tied to the lifetime of the runner pod.
When a runner pod exits or crashes while a workflow pod is still running,
then the runner gets stuck waiting indefinitely for storage.
By cleaning up the workflow pod and thus detaching the volume,
the runner can become available again.
