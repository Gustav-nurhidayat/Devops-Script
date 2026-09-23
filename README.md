# SimpleJourney DevOps Technical Test

Technical test submission for the Shift Engineer - DevOps position at PT Simple Journey Indonesia.

## Overview

This project demonstrates a basic CI/CD workflow for a Go HTTP application using:

- Go
- Docker
- Jenkins
- Git
- Docker-based hotfix deployment

The pipeline performs:

1. Source checkout from GitHub
2. Unit testing
3. Version generation from Git commit
4. Multi-stage Docker image build
5. Image verification
6. Hotfix deployment to an existing running container
7. Deployment health check

---

# Project Structure

simplejourney-devops-test/
├── Dockerfile
├── Jenkinsfile
├── README.md
├── go.mod
├── main.go
├── main_test.go
├── jenkins/
│   └── Dockerfile
├── scripts/
│   └── deploy-hotfix.sh
└── supervisor/
    └── main.go


## Application

The application exposes:


GET /


Example response:

Hello, DevOps! version=dev

The application version is injected during compilation using Go `-ldflags`.

Example:


go build \
  -ldflags="-X main.version=1.0.0" \
  -o app .


---

# Part I — Docker

## 1. Multi-stage Docker Build

The project uses a multi-stage Dockerfile.

The builder stage uses:

golang:1.26-alpine

The final stage uses:

scratch

The Go application and supervisor are compiled as static Linux binaries:


CGO_ENABLED=0
GOOS=linux
GOARCH=amd64


The final image contains only the compiled binaries and does not contain the Go toolchain or build dependencies.

### Final Runtime Image

The final stage uses:


FROM scratch

with:


ENTRYPOINT ["/supervisor"]

This keeps the runtime image small and minimizes unnecessary packages.

## 2. Version Injection

The application version is injected using:


-ldflags="-s -w -X main.version=${VERSION}"


Example:

docker build \
  --build-arg VERSION=1.0.0 \
  -t simplejourney-devops:1.0.0 .


The application then returns:


Hello, DevOps! version=1.0.0


## 3. Image Size

The Docker image is approximately:


7.7 MB


The image contains the statically compiled application and supervisor without the Go toolchain or build dependencies.

---

# Part II — Running Container and Hotfix Deployment

## 4. Run the Application

Example:

docker run -d \
  --name simplejourney \
  -p 8080:8080 \
  --restart unless-stopped \
  simplejourney-devops:1.0.0


Test:


curl http://localhost:8080/


Expected:


Hello, DevOps! version=1.0.0


The `--restart unless-stopped` policy allows Docker to automatically restart the container after a failure or Docker daemon restart.

---

# 5. Hotfix Without Rebuilding the Docker Image

The hotfix mechanism allows the application binary to be replaced inside an already running container.

The existing Docker image and container are not removed or rebuilt.

The workflow is:


Build binary on host/CI
        ↓
docker cp
        ↓
/app.new
        ↓
SIGHUP
        ↓
Supervisor replaces /app
        ↓
Application restarts
        ↓
Health check


The deployment script is:


scripts/deploy-hotfix.sh


Example:


VERSION=1.0.3 \
CONTAINER=simplejourney-hotfix \
PORT=8081 \
./scripts/deploy-hotfix.sh


The binary is copied into the running container:


docker cp deploy/app simplejourney-hotfix:/app.new


The supervisor is then notified:


docker kill --signal=SIGHUP simplejourney-hotfix


The supervisor replaces the running binary and starts the new version.

---

# 6. Hotfix Verification

Example deployment output:

==> Deploying version: 1.0.3
==> Target container: simplejourney-hotfix
==> Building hotfix binary...
==> Copying binary into running container...
==> Triggering hot reload...
==> Waiting for application...
==> Deployment successful.
Hello, DevOps! version=1.0.3


The container itself remains running while the application binary is replaced.

The mechanism introduces a short interruption because the supervisor stops the old process before starting the new process.

This approach is useful for a simple production hotfix because it avoids rebuilding and replacing the existing container image. The trade-off is that it modifies the runtime container directly and is therefore less reproducible than deploying a new immutable image.

---

# Part III — Jenkins CI/CD

## 7. Source Checkout

Jenkins checks out the source code from the GitHub repository using a Jenkins-managed GitHub credential.

The GitHub Personal Access Token is stored in Jenkins Credentials and is not hardcoded in the Jenkinsfile.

---

## 8. Automated Tests

The pipeline runs:


go test ./...


The pipeline must pass this stage before continuing to the Docker build and deployment stages.

Example successful result:


ok   simplejourney-devops-test
?    simplejourney-devops-test/supervisor [no test files]


A failing test causes the pipeline to stop before deployment.

---

## 9. Version From Git Commit

The Jenkins pipeline obtains the short Git commit hash:


git rev-parse --short HEAD


The value is then used as the application version.

Example:


651f6ba

The Docker image is tagged using the same version:


simplejourney-devops:651f6ba


The application is compiled using:


-ldflags="-s -w -X main.version=651f6ba"


This provides traceability between the deployed application and the source commit.

---

# 10. Docker Image Build

Jenkins executes:


docker build \
  --build-arg VERSION="${VERSION}" \
  -t "${IMAGE_NAME}:${VERSION}" \
  .


The image is then verified:


docker image inspect "${IMAGE_NAME}:${VERSION}"


Example:


simplejourney-devops:651f6ba


Image size:


7.7 MB


No external container registry is required for this test environment. Jenkins accesses the Docker daemon through the Docker socket.

---

# 11. Automated Hotfix Deployment

After the image has been successfully built and verified, Jenkins executes:


VERSION="${VERSION}" \
CONTAINER="${TARGET_CONTAINER}" \
PORT="${TARGET_PORT}" \
./scripts/deploy-hotfix.sh


The deployment target is:


Container: simplejourney-hotfix
Port: 8081


The deployment script:

1. Verifies the target container exists.
2. Verifies the container is running.
3. Builds the new static binary.
4. Copies the binary to `/app.new`.
5. Sends `SIGHUP` to the supervisor.
6. Waits for the new version.
7. Performs an HTTP health check.
8. Fails the pipeline if the expected version is not returned.

---

# Jenkins Pipeline Stages


Checkout
   ↓
Test
   ↓
Prepare Version
   ↓
Build Docker Image
   ↓
Verify Image
   ↓
Deploy Hotfix


The successful Jenkins Build #4 produced:


Build version: 651f6ba


The Docker image was successfully built:


Successfully tagged simplejourney-devops:651f6ba


The deployment completed successfully:


==> Deployment successful.
Hello, DevOps! version=651f6ba

Final Jenkins result:


Finished: SUCCESS


---

# Jenkins Environment

Jenkins runs inside Docker using a custom Jenkins image.

The Jenkins image includes:

* Jenkins LTS
* JDK 21
* Go 1.26.2
* Git
* Docker CLI

The Jenkins container has access to the host Docker daemon through:


/var/run/docker.sock


This allows Jenkins to build Docker images and interact with the deployment container.

---

# Failure Handling

The pipeline uses shell commands with:


set -e


Therefore, a failed command causes the corresponding stage to fail.

For example, if:


go test ./...


fails, the pipeline does not continue to:


Build Docker Image


or:

Deploy Hotfix


The deployment script also performs an HTTP health check after deployment.

If the expected version is not returned, the deployment stage fails.

---

# Rollback

If a deployment needs to be rolled back, the previous known-good application binary can be rebuilt from the corresponding Git commit and copied into the running container using the same hotfix mechanism.

Example:


git checkout <known-good-commit>

VERSION=<previous-version> \
CONTAINER=simplejourney-hotfix \
PORT=8081 \
./scripts/deploy-hotfix.sh


This restores the application binary without rebuilding or replacing the running container.

For a production environment, an immutable image-based deployment with versioned images and automated rollback would provide stronger reproducibility.

---

# Security Considerations

* GitHub credentials are stored in Jenkins Credentials.
* Secrets are not hardcoded in the Jenkinsfile.
* The runtime image does not contain the Go compiler or build dependencies.
* The final image uses `scratch` to minimize the runtime attack surface.
* Docker build and deployment are performed through the Jenkins-controlled environment.

---

# Tools and Versions

| Tool                 | Version |
| -------------------- | ------- |
| Go                   | 1.26.2  |
| Docker Engine (host) | 28.5.2  |
| Docker CLI (Jenkins) | 29.8.1  |
| Jenkins              | LTS     |
| Git (Jenkins)        | 2.47.3  |
| Runtime Image        | scratch |

---

# Evidence

The successful Jenkins pipeline demonstrates:

* Source checkout from GitHub
* Automated Go tests
* Commit-based versioning
* Docker image build
* Image verification
* Automated hotfix deployment
* HTTP health check
* Successful end-to-end pipeline

Latest successful pipeline:

Jenkins Build #4
Commit: 651f6ba
Image: simplejourney-devops:651f6ba
Deployed version: 651f6ba
Result: SUCCESS


---

# Conclusion

This project demonstrates a complete basic CI/CD workflow using Go, Docker, Jenkins, and Git.

The pipeline validates the source code, creates a versioned Docker image, verifies the image, and automatically deploys the application using a hotfix mechanism without replacing the existing running container.

The implementation is intended as a technical-test demonstration and can be extended in a production environment with an external container registry, immutable deployments, automated rollback, secrets management, monitoring, and zero-downtime deployment strategies.
