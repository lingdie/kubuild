/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BuildTaskPhase is a coarse-grained lifecycle phase for a build.
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
type BuildTaskPhase string

const (
	BuildTaskPhasePending   BuildTaskPhase = "Pending"
	BuildTaskPhaseRunning   BuildTaskPhase = "Running"
	BuildTaskPhaseSucceeded BuildTaskPhase = "Succeeded"
	BuildTaskPhaseFailed    BuildTaskPhase = "Failed"
)

// BuildContextType selects the source type of the build context.
// +kubebuilder:validation:Enum=Git;PVC;S3
type BuildContextType string

const (
	BuildContextTypeGit BuildContextType = "Git"
	BuildContextTypePVC BuildContextType = "PVC"
	BuildContextTypeS3  BuildContextType = "S3"
)

// DockerfileSourceType selects how Dockerfile is provided.
// +kubebuilder:validation:Enum=Path;Inline
type DockerfileSourceType string

const (
	DockerfileSourceTypePath   DockerfileSourceType = "Path"
	DockerfileSourceTypeInline DockerfileSourceType = "Inline"
)

// LocalSecretReference references a Secret in the same namespace.
type LocalSecretReference struct {
	// name is the name of the Secret.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// BuildContext defines the source of the build context.
// Exactly one of git/pvc/s3 must be set, according to type.
// +kubebuilder:validation:XValidation:rule="self.type == 'Git' ? has(self.git) && !has(self.pvc) && !has(self.s3) : true",message="when context.type=Git, context.git must be set and pvc/s3 must be empty"
// +kubebuilder:validation:XValidation:rule="self.type == 'PVC' ? has(self.pvc) && !has(self.git) && !has(self.s3) : true",message="when context.type=PVC, context.pvc must be set and git/s3 must be empty"
// +kubebuilder:validation:XValidation:rule="self.type == 'S3' ? has(self.s3) && !has(self.git) && !has(self.pvc) : true",message="when context.type=S3, context.s3 must be set and git/pvc must be empty"
type BuildContext struct {
	// type selects which context source to use.
	// +kubebuilder:validation:Required
	Type BuildContextType `json:"type"`

	// git defines the git repository context.
	// +optional
	Git *GitContext `json:"git,omitempty"`

	// pvc defines the PVC-based context.
	// +optional
	PVC *PVCContext `json:"pvc,omitempty"`

	// s3 defines an S3/OSS object-based context (e.g. tarball).
	// +optional
	S3 *S3Context `json:"s3,omitempty"`
}

type GitContext struct {
	// url is the git repo URL.
	// +kubebuilder:validation:MinLength=1
	URL string `json:"url"`

	// revision is a branch, tag, or commit sha.
	// +optional
	Revision string `json:"revision,omitempty"`

	// subPath is a path inside the repo to use as build context.
	// +optional
	SubPath string `json:"subPath,omitempty"`

	// secretRef optionally references credentials (e.g. SSH key, token) for private repos.
	// The Secret must be in the same namespace as the BuildTask.
	// +optional
	SecretRef *LocalSecretReference `json:"secretRef,omitempty"`

	// depth optionally performs a shallow clone.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Depth *int32 `json:"depth,omitempty"`
}

type PVCContext struct {
	// claimName is the PVC name.
	// +kubebuilder:validation:MinLength=1
	ClaimName string `json:"claimName"`

	// path is a path inside the mounted PVC to use as build context.
	// +optional
	// +kubebuilder:default="/"
	Path string `json:"path,omitempty"`
}

type S3Context struct {
	// bucket is the bucket name.
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`

	// key is the object key (e.g. a tarball containing build context).
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// endpoint is the S3-compatible endpoint (optional).
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// region is the region (optional).
	// +optional
	Region string `json:"region,omitempty"`

	// secretRef references credentials for S3/OSS access (required for private buckets).
	// The Secret must be in the same namespace as the BuildTask.
	// +optional
	SecretRef *LocalSecretReference `json:"secretRef,omitempty"`
}

// DockerfileSource defines how Dockerfile is provided.
// +kubebuilder:validation:XValidation:rule="self.type == 'Path' ? (self.path != ” && self.inline == ”) : true",message="when dockerfile.type=Path, dockerfile.path must be non-empty and dockerfile.inline must be empty"
// +kubebuilder:validation:XValidation:rule="self.type == 'Inline' ? (self.inline != ”) : true",message="when dockerfile.type=Inline, dockerfile.inline must be non-empty"
type DockerfileSource struct {
	// type selects path or inline Dockerfile.
	// +kubebuilder:validation:Required
	Type DockerfileSourceType `json:"type"`

	// path is a path within the build context (used when type=Path).
	// +optional
	// +kubebuilder:default="Dockerfile"
	Path string `json:"path,omitempty"`

	// inline is the inline Dockerfile content (used when type=Inline).
	// +optional
	Inline string `json:"inline,omitempty"`
}

type BuildOutput struct {
	// push controls whether to push built image(s) to the registry.
	// +optional
	// +kubebuilder:default=true
	Push *bool `json:"push,omitempty"`

	// images are additional image references (tags) to push, besides spec.image.
	// +optional
	Images []string `json:"images,omitempty"`

	// skipTLSVerify skips TLS verification when talking to registries.
	// +optional
	SkipTLSVerify *bool `json:"skipTLSVerify,omitempty"`

	// insecure allows plain HTTP registry (not recommended).
	// +optional
	Insecure *bool `json:"insecure,omitempty"`
}

// BuildahSpec contains buildah runtime options.
// This API is intentionally rootless-only and does not expose privileged mode.
// +kubebuilder:validation:XValidation:rule="!has(self.rootless) || self.rootless == true",message="buildah.rootless must be true; privileged mode is not supported"
type BuildahSpec struct {
	// rootless indicates the job should run rootless.
	// Must be true for this controller.
	// +optional
	// +kubebuilder:default=true
	Rootless *bool `json:"rootless,omitempty"`

	// storageDriver configures containers/storage driver (common values: overlay, vfs).
	// +optional
	StorageDriver string `json:"storageDriver,omitempty"`

	// env optionally sets extra environment variables for buildah container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type CacheSpec struct {
	// enabled enables build cache.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// pvcName references an existing PVC for cache storage (same namespace).
	// +optional
	PVCName string `json:"pvcName,omitempty"`

	// subPath is an optional path within the PVC.
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

type RetentionSpec struct {
	// successfulJobsTTLSecondsAfterFinished sets TTL for successful jobs.
	// +optional
	// +kubebuilder:validation:Minimum=0
	SuccessfulJobsTTLSecondsAfterFinished *int32 `json:"successfulJobsTTLSecondsAfterFinished,omitempty"`

	// failedJobsTTLSecondsAfterFinished sets TTL for failed jobs.
	// +optional
	// +kubebuilder:validation:Minimum=0
	FailedJobsTTLSecondsAfterFinished *int32 `json:"failedJobsTTLSecondsAfterFinished,omitempty"`
}

// TriggerSpec defines how a build is triggered.
// This controller uses a manual trigger nonce: update nonce to start a new build.
type TriggerSpec struct {
	// manual trigger; changing nonce requests a new build.
	// +optional
	Manual *ManualTrigger `json:"manual,omitempty"`
}

type ManualTrigger struct {
	// nonce is an arbitrary string. Changing it triggers a new build.
	// Common patterns: unix timestamp, git sha, or a UUID.
	// +optional
	Nonce string `json:"nonce,omitempty"`
}

// BuildTaskSpec defines the desired state of BuildTask
type BuildTaskSpec struct {
	// image is the primary target image reference (e.g. registry/repo:tag).
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// output controls push behavior and extra tags.
	// +optional
	Output *BuildOutput `json:"output,omitempty"`

	// context defines where the build context comes from.
	// +kubebuilder:validation:Required
	Context BuildContext `json:"context"`

	// dockerfile defines where Dockerfile is loaded from.
	// +optional
	Dockerfile *DockerfileSource `json:"dockerfile,omitempty"`

	// buildArgs are build-time args.
	// +optional
	BuildArgs map[string]string `json:"buildArgs,omitempty"`

	// pushSecretRef references docker registry credentials (kubernetes.io/dockerconfigjson).
	// The Secret must be in the same namespace as the BuildTask.
	// +optional
	PushSecretRef *LocalSecretReference `json:"pushSecretRef,omitempty"`

	// serviceAccountName is the ServiceAccount used by the Job.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// resources defines compute resource requests/limits for the build container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// timeoutSeconds limits the max duration of a build (mapped to Job activeDeadlineSeconds).
	// +optional
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds *int64 `json:"timeoutSeconds,omitempty"`

	// backoffLimit is the Job backoffLimit (retries on failure).
	// +optional
	// +kubebuilder:validation:Minimum=0
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// retention defines TTL policy for successful/failed Jobs created by this BuildTask.
	// +optional
	Retention *RetentionSpec `json:"retention,omitempty"`

	// cache configures optional cache storage.
	// +optional
	Cache *CacheSpec `json:"cache,omitempty"`

	// buildah configures rootless buildah options.
	// +optional
	Buildah *BuildahSpec `json:"buildah,omitempty"`

	// trigger controls manual trigger behavior.
	// +optional
	Trigger *TriggerSpec `json:"trigger,omitempty"`
}

// BuildTaskStatus defines the observed state of BuildTask.
type BuildTaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the BuildTask resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// phase is a coarse-grained state machine for the build.
	// +optional
	Phase BuildTaskPhase `json:"phase,omitempty"`

	// jobName is the name of the Job created for this build.
	// +optional
	JobName string `json:"jobName,omitempty"`

	// podName is the latest Pod name of the Job (best-effort, may change across retries).
	// +optional
	PodName string `json:"podName,omitempty"`

	// startTime is when the build started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// endTime is when the build completed.
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// imageDigest is the resulting image digest (e.g. sha256:...).
	// +optional
	ImageDigest string `json:"imageDigest,omitempty"`

	// lastTriggerNonce is the last observed trigger nonce.
	// +optional
	LastTriggerNonce string `json:"lastTriggerNonce,omitempty"`

	// logURL is an optional URL pointing to build logs.
	// +optional
	LogURL string `json:"logURL,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image"
// +kubebuilder:printcolumn:name="Job",type="string",JSONPath=".status.jobName"

// BuildTask is the Schema for the buildtasks API
type BuildTask struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of BuildTask
	// +required
	Spec BuildTaskSpec `json:"spec"`

	// status defines the observed state of BuildTask
	// +optional
	Status BuildTaskStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// BuildTaskList contains a list of BuildTask
type BuildTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []BuildTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BuildTask{}, &BuildTaskList{})
}
