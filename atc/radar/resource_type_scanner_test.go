package radar_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/creds"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/db/dbfakes"
	"github.com/concourse/concourse/atc/db/lock"
	"github.com/concourse/concourse/atc/db/lock/lockfakes"
	. "github.com/concourse/concourse/atc/radar"
	"github.com/concourse/concourse/atc/worker"
	"github.com/concourse/concourse/atc/worker/workerfakes"

	rfakes "github.com/concourse/concourse/atc/resource/resourcefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourceTypeScanner", func() {
	var (
		epoch time.Time

		fakeConn *dbfakes.FakeConn
		fakeTx   *dbfakes.FakeTx

		fakeContainer       *workerfakes.FakeContainer
		fakeWorker          *workerfakes.FakeWorker
		fakePool            *workerfakes.FakePool
		fakeStrategy        *workerfakes.FakeContainerPlacementStrategy
		fakeResourceFactory *rfakes.FakeResourceFactory
		fakeDBPipeline      *dbfakes.FakePipeline
		fakeResourceConfig  *dbfakes.FakeResourceConfig
		fakeClock           *fakeclock.FakeClock
		interval            time.Duration
		variables           creds.Variables

		fakeResourceType      *dbfakes.FakeResourceType
		versionedResourceType atc.VersionedResourceType

		scanner Scanner

		fakeLock *lockfakes.FakeLock
		teamID   = 123
	)

	BeforeEach(func() {
		fakeLock = &lockfakes.FakeLock{}
		interval = 1 * time.Minute
		variables = template.StaticVariables{
			"source-params": "some-secret-sauce",
		}

		versionedResourceType = atc.VersionedResourceType{
			ResourceType: atc.ResourceType{
				Name:   "some-custom-resource",
				Type:   "registry-image",
				Source: atc.Source{"custom": "((source-params))"},
				Tags:   atc.Tags{"some-tag"},
			},
			Version: atc.Version{"custom": "version"},
		}

		fakeTx = new(dbfakes.FakeTx)
		fakeTx.RollbackReturns(nil)
		fakeConn = new(dbfakes.FakeConn)
		fakeConn.BeginReturns(fakeTx, nil)

		fakeClock = fakeclock.NewFakeClock(epoch)

		fakeContainer = new(workerfakes.FakeContainer)
		fakeStrategy = new(workerfakes.FakeContainerPlacementStrategy)
		fakePool = new(workerfakes.FakePool)
		fakeWorker = new(workerfakes.FakeWorker)
		fakeResourceFactory = new(rfakes.FakeResourceFactory)
		fakeResourceType = new(dbfakes.FakeResourceType)
		fakeDBPipeline = new(dbfakes.FakePipeline)
		fakeResourceConfig = new(dbfakes.FakeResourceConfig)
		fakeResourceConfig.IDReturns(123)

		fakeResourceType.IDReturns(39)
		fakeResourceType.NameReturns("some-custom-resource")
		fakeResourceType.TypeReturns("registry-image")
		fakeResourceType.SourceReturns(atc.Source{"custom": "((source-params))"})
		fakeResourceType.VersionReturns(atc.Version{"custom": "version"}, nil)
		fakeResourceType.TagsReturns(atc.Tags{"some-tag"})
		fakeResourceType.SetResourceConfigReturns(fakeResourceConfig, nil)

		fakeDBPipeline.IDReturns(42)
		fakeDBPipeline.NameReturns("some-pipeline")
		fakeDBPipeline.TeamIDReturns(teamID)
		fakeDBPipeline.ReloadReturns(true, nil)
		fakeDBPipeline.ResourceTypesReturns([]db.ResourceType{fakeResourceType}, nil)
		fakeDBPipeline.ResourceTypeReturns(fakeResourceType, true, nil)

		scanner = NewResourceTypeScanner(
			fakeConn,
			fakeClock,
			fakePool,
			fakeResourceFactory,
			interval,
			fakeDBPipeline,
			"https://www.example.com",
			variables,
			fakeStrategy,
		)
	})

	Describe("Run", func() {
		var (
			fakeResource   *rfakes.FakeResource
			actualInterval time.Duration
			runErr         error
		)

		BeforeEach(func() {
			fakeWorker.NameReturns("some-worker")
			fakePool.FindOrChooseWorkerForContainerReturns(fakeWorker, nil)

			fakeContainer.HandleReturns("some-handle")
			fakeWorker.FindOrCreateContainerReturns(fakeContainer, nil)

			fakeResource = new(rfakes.FakeResource)
			fakeResourceFactory.NewResourceForContainerReturns(fakeResource, nil)
		})

		JustBeforeEach(func() {
			actualInterval, runErr = scanner.Run(lagertest.NewTestLogger("test"), fakeResourceType.Name())
		})

		Context("when the lock cannot be acquired", func() {
			BeforeEach(func() {
				fakeResourceConfig.AcquireResourceCheckingLockReturns(nil, false, nil)
			})

			It("does not check", func() {
				Expect(fakeResource.CheckCallCount()).To(Equal(0))
			})

			It("returns the configured interval", func() {
				Expect(runErr).To(Equal(ErrFailedToAcquireLock))
				Expect(actualInterval).To(Equal(interval))
			})
		})

		Context("when the lock can be acquired", func() {
			BeforeEach(func() {
				fakeResourceConfig.AcquireResourceCheckingLockReturns(fakeLock, true, nil)
			})

			Context("when the last checked was not updated", func() {
				BeforeEach(func() {
					fakeResourceConfig.UpdateLastCheckedReturns(false, nil)
				})

				It("does not check", func() {
					Expect(fakeResource.CheckCallCount()).To(Equal(0))
				})

				It("returns the configured interval", func() {
					Expect(runErr).To(Equal(ErrFailedToAcquireLock))
					Expect(actualInterval).To(Equal(interval))
				})
			})

			Context("when the last checked was updated", func() {
				BeforeEach(func() {
					fakeResourceConfig.UpdateLastCheckedReturns(true, nil)
				})

				It("checks immediately", func() {
					Expect(fakeResource.CheckCallCount()).To(Equal(1))
				})

				It("constructs the resource of the correct type", func() {
					Expect(fakeResourceType.SetResourceConfigCallCount()).To(Equal(1))
					_, resourceSource, resourceTypes := fakeResourceType.SetResourceConfigArgsForCall(0)
					Expect(resourceSource).To(Equal(atc.Source{"custom": "some-secret-sauce"}))
					Expect(resourceTypes).To(Equal(creds.VersionedResourceTypes{}))

					_, owner, containerSpec, workerSpec, _ := fakePool.FindOrChooseWorkerForContainerArgsForCall(0)
					Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
					Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
						ResourceType: "registry-image",
					}))
					Expect(containerSpec.Tags).To(Equal([]string{"some-tag"}))
					Expect(containerSpec.TeamID).To(Equal(123))
					Expect(workerSpec).To(Equal(worker.WorkerSpec{
						ResourceType:  "registry-image",
						Tags:          []string{"some-tag"},
						ResourceTypes: creds.VersionedResourceTypes{},
						TeamID:        123,
					}))

					Expect(fakeWorker.FindOrCreateContainerCallCount()).To(Equal(1))
					_, _, _, owner, metadata, containerSpec, resourceTypes := fakeWorker.FindOrCreateContainerArgsForCall(0)
					Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
					Expect(metadata).To(Equal(db.ContainerMetadata{
						Type: db.ContainerTypeCheck,
					}))
					Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
						ResourceType: "registry-image",
					}))
					Expect(containerSpec.Tags).To(Equal([]string{"some-tag"}))
					Expect(containerSpec.TeamID).To(Equal(123))
					Expect(resourceTypes).To(Equal(creds.VersionedResourceTypes{}))
				})

				Context("when the resource type overrides a base resource type", func() {
					BeforeEach(func() {
						otherResourceType := fakeResourceType

						fakeResourceType = new(dbfakes.FakeResourceType)
						fakeResourceType.IDReturns(40)
						fakeResourceType.NameReturns("registry-image")
						fakeResourceType.TypeReturns("registry-image")
						fakeResourceType.SourceReturns(atc.Source{"custom": "((source-params))"})
						fakeResourceType.VersionReturns(atc.Version{"custom": "image-version"}, nil)
						fakeResourceType.SetResourceConfigReturns(fakeResourceConfig, nil)

						fakeDBPipeline.ResourceTypesReturns([]db.ResourceType{
							fakeResourceType,
							otherResourceType,
						}, nil)
						fakeDBPipeline.ResourceTypeReturns(fakeResourceType, true, nil)
					})

					It("constructs the resource of the correct type", func() {
						Expect(fakeResourceType.SetResourceConfigCallCount()).To(Equal(1))
						_, resourceSource, resourceTypes := fakeResourceType.SetResourceConfigArgsForCall(0)
						Expect(resourceSource).To(Equal(atc.Source{"custom": "some-secret-sauce"}))
						Expect(resourceTypes).To(Equal(creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{
							versionedResourceType,
						})))

						Expect(fakeResourceType.SetCheckSetupErrorCallCount()).To(Equal(1))
						err := fakeResourceType.SetCheckSetupErrorArgsForCall(0)
						Expect(err).To(BeNil())

						_, owner, containerSpec, workerSpec, _ := fakePool.FindOrChooseWorkerForContainerArgsForCall(0)
						Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
						Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
							ResourceType: "registry-image",
						}))
						Expect(containerSpec.TeamID).To(Equal(123))
						Expect(workerSpec).To(Equal(worker.WorkerSpec{
							ResourceType:  "registry-image",
							ResourceTypes: creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{versionedResourceType}),
							TeamID:        123,
						}))

						Expect(fakeWorker.FindOrCreateContainerCallCount()).To(Equal(1))
						_, _, _, owner, metadata, containerSpec, resourceTypes := fakeWorker.FindOrCreateContainerArgsForCall(0)
						Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
						Expect(metadata).To(Equal(db.ContainerMetadata{
							Type: db.ContainerTypeCheck,
						}))
						Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
							ResourceType: "registry-image",
						}))
						Expect(containerSpec.TeamID).To(Equal(123))
						Expect(resourceTypes).To(Equal(creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{
							versionedResourceType,
						})))
					})
				})

				Context("when the resource type config has a specified check interval", func() {
					BeforeEach(func() {
						fakeResourceType.CheckEveryReturns("10ms")
						fakeDBPipeline.ResourceTypeReturns(fakeResourceType, true, nil)
					})

					It("leases for the configured interval", func() {
						Expect(fakeResourceConfig.AcquireResourceCheckingLockCallCount()).To(Equal(1))
						Expect(fakeResourceConfig.UpdateLastCheckedCallCount()).To(Equal(1))

						leaseInterval, immediate := fakeResourceConfig.UpdateLastCheckedArgsForCall(0)
						Expect(leaseInterval).To(Equal(10 * time.Millisecond))
						Expect(immediate).To(BeFalse())

						Eventually(fakeLock.ReleaseCallCount()).Should(Equal(1))
					})

					It("returns configured interval", func() {
						Expect(actualInterval).To(Equal(10 * time.Millisecond))
					})

					Context("when the interval cannot be parsed", func() {
						BeforeEach(func() {
							fakeResourceType.CheckEveryReturns("bad-value")
							fakeDBPipeline.ResourceTypeReturns(fakeResourceType, true, nil)
						})

						It("sets the check error", func() {
							Expect(fakeResourceType.SetCheckSetupErrorCallCount()).To(Equal(1))

							resourceErr := fakeResourceType.SetCheckSetupErrorArgsForCall(0)
							Expect(resourceErr).To(MatchError("time: invalid duration bad-value"))
						})

						It("returns an error", func() {
							Expect(runErr).To(HaveOccurred())
						})
					})
				})

				It("grabs a periodic resource checking lock before checking, breaks lock after done", func() {
					Expect(fakeResourceConfig.AcquireResourceCheckingLockCallCount()).To(Equal(1))
					Expect(fakeResourceConfig.UpdateLastCheckedCallCount()).To(Equal(1))

					leaseInterval, immediate := fakeResourceConfig.UpdateLastCheckedArgsForCall(0)
					Expect(leaseInterval).To(Equal(interval))
					Expect(immediate).To(BeFalse())

					Eventually(fakeLock.ReleaseCallCount()).Should(Equal(1))
				})

				Context("when there is no current version", func() {
					BeforeEach(func() {
						fakeResourceType.VersionReturns(nil, nil)
					})

					It("checks from nil", func() {
						_, _, _, version := fakeResource.CheckArgsForCall(0)
						Expect(version).To(Equal(map[atc.Space]atc.Version{}))
					})
				})

				Context("when there are current versions", func() {
					BeforeEach(func() {
						fakeResourceVersion1 := new(dbfakes.FakeResourceVersion)
						fakeResourceVersion1.IDReturns(1)
						fakeResourceVersion1.VersionReturns(db.Version{"version": "1"})
						fakeResourceVersion1.SpaceReturns(atc.Space("space1"))

						fakeResourceVersion2 := new(dbfakes.FakeResourceVersion)
						fakeResourceVersion2.IDReturns(2)
						fakeResourceVersion2.VersionReturns(db.Version{"version": "2"})
						fakeResourceVersion2.SpaceReturns(atc.Space("space2"))

						fakeResourceConfig.LatestVersionsReturns([]db.ResourceVersion{fakeResourceVersion1, fakeResourceVersion2}, nil)
					})

					It("checks with them", func() {
						_, checkHandler, _, version := fakeResource.CheckArgsForCall(0)
						var expectedCheckHandler CheckEventHandler
						Expect(checkHandler).To(BeAssignableToTypeOf(&expectedCheckHandler))
						Expect(version).To(Equal(map[atc.Space]atc.Version{
							atc.Space("space1"): atc.Version{"version": "1"},
							atc.Space("space2"): atc.Version{"version": "2"},
						}))
					})
				})

				Context("when checking fails", func() {
					disaster := errors.New("nope")

					BeforeEach(func() {
						fakeResource.CheckReturns(disaster)
					})

					It("exits with the failure", func() {
						Expect(runErr).To(HaveOccurred())
						Expect(runErr).To(Equal(disaster))
					})

					It("sets the resource's check error", func() {
						Expect(fakeResourceConfig.SetCheckErrorCallCount()).To(Equal(1))

						err := fakeResourceConfig.SetCheckErrorArgsForCall(0)
						Expect(err).To(Equal(disaster))
					})
				})

				Context("when the pipeline is paused", func() {
					BeforeEach(func() {
						fakeDBPipeline.CheckPausedReturns(true, nil)
					})

					It("does not check", func() {
						Expect(fakeResource.CheckCallCount()).To(BeZero())
					})

					It("returns the default interval", func() {
						Expect(actualInterval).To(Equal(interval))
					})

					It("does not return an error", func() {
						Expect(runErr).NotTo(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Scan", func() {
		var (
			fakeResource *rfakes.FakeResource
			runErr       error
		)

		BeforeEach(func() {
			fakeWorker.NameReturns("some-worker")
			fakePool.FindOrChooseWorkerForContainerReturns(fakeWorker, nil)

			fakeContainer.HandleReturns("some-handle")
			fakeWorker.FindOrCreateContainerReturns(fakeContainer, nil)

			fakeResource = new(rfakes.FakeResource)
			fakeResourceFactory.NewResourceForContainerReturns(fakeResource, nil)
		})

		JustBeforeEach(func() {
			runErr = scanner.Scan(lagertest.NewTestLogger("test"), fakeResourceType.Name())
		})

		Context("when the lock can be acquired and last checked is updated", func() {
			BeforeEach(func() {
				fakeResourceConfig.AcquireResourceCheckingLockReturns(fakeLock, true, nil)
				fakeResourceConfig.UpdateLastCheckedReturns(true, nil)
			})

			It("checks immediately", func() {
				Expect(fakeResource.CheckCallCount()).To(Equal(1))
			})

			It("constructs the resource of the correct type", func() {
				Expect(fakeResourceType.SetResourceConfigCallCount()).To(Equal(1))
				_, resourceSource, resourceTypes := fakeResourceType.SetResourceConfigArgsForCall(0)
				Expect(resourceSource).To(Equal(atc.Source{"custom": "some-secret-sauce"}))
				Expect(resourceTypes).To(Equal(creds.VersionedResourceTypes{}))

				Expect(fakeResourceType.SetCheckSetupErrorCallCount()).To(Equal(1))
				err := fakeResourceType.SetCheckSetupErrorArgsForCall(0)
				Expect(err).To(BeNil())

				_, owner, containerSpec, workerSpec, _ := fakePool.FindOrChooseWorkerForContainerArgsForCall(0)
				Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
				Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
					ResourceType: "registry-image",
				}))
				Expect(containerSpec.Tags).To(Equal([]string{"some-tag"}))
				Expect(containerSpec.TeamID).To(Equal(123))
				Expect(workerSpec).To(Equal(worker.WorkerSpec{
					ResourceType:  "registry-image",
					Tags:          []string{"some-tag"},
					ResourceTypes: creds.VersionedResourceTypes{},
					TeamID:        123,
				}))

				Expect(fakeWorker.FindOrCreateContainerCallCount()).To(Equal(1))
				_, _, _, owner, metadata, containerSpec, resourceTypes := fakeWorker.FindOrCreateContainerArgsForCall(0)
				Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
				Expect(metadata).To(Equal(db.ContainerMetadata{
					Type: db.ContainerTypeCheck,
				}))
				Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
					ResourceType: "registry-image",
				}))
				Expect(containerSpec.Tags).To(Equal([]string{"some-tag"}))
				Expect(containerSpec.TeamID).To(Equal(123))
				Expect(resourceTypes).To(Equal(creds.VersionedResourceTypes{}))
			})

			Context("when the resource type depends on another custom type", func() {
				var otherResourceType *dbfakes.FakeResourceType

				BeforeEach(func() {
					otherResourceType = new(dbfakes.FakeResourceType)
					otherResourceType.IDReturns(39)
					otherResourceType.NameReturns("custom-resource-parent")
					otherResourceType.TypeReturns("registry-image")

					fakeResourceType = new(dbfakes.FakeResourceType)
					fakeResourceType.IDReturns(40)
					fakeResourceType.NameReturns("custom-resource-child")
					fakeResourceType.TypeReturns("custom-resource-parent")
					fakeResourceType.SetResourceConfigReturns(fakeResourceConfig, nil)

					// testing recursion is fun!
					fakeDBPipeline.ResourceTypesReturnsOnCall(0, []db.ResourceType{
						fakeResourceType,
						otherResourceType,
					}, nil)
					fakeDBPipeline.ResourceTypeReturnsOnCall(0, fakeResourceType, true, nil)
				})

				Context("when the custom type does not yet have a version", func() {
					BeforeEach(func() {
						otherResourceType.VersionReturns(nil, nil)
					})

					It("checks for versions of the parent resource type", func() {
						By("calling .scan() for the resource type name")
						Expect(fakeDBPipeline.ResourceTypeCallCount()).To(Equal(2))
						Expect(fakeDBPipeline.ResourceTypeArgsForCall(1)).To(Equal("custom-resource-parent"))
					})

					Context("when the check for the parent succeeds", func() {
						It("reloads the resource types from the database", func() {
							Expect(runErr).ToNot(HaveOccurred())
							Expect(fakeDBPipeline.ResourceTypesCallCount()).To(Equal(4))
						})
					})

					Context("something fails in the parent resource scan", func() {
						var parentResourceTypeErr = errors.New("jma says no recursion in production")
						BeforeEach(func() {
							fakeDBPipeline.ResourceTypeReturnsOnCall(1, otherResourceType, true, parentResourceTypeErr)
						})

						It("returns the error from scanning the parent", func() {
							Expect(runErr).To(Equal(parentResourceTypeErr))
						})

						It("saves the error to check_error on resource type row in db", func() {
							Expect(fakeResourceType.SetCheckSetupErrorCallCount()).To(Equal(1))

							err := fakeResourceType.SetCheckSetupErrorArgsForCall(0)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(Equal("jma says no recursion in production"))
						})
					})
				})

				Context("when the custom type has a version already", func() {
					BeforeEach(func() {
						otherResourceType.VersionReturns(atc.Version{"custom": "image-version"}, nil)
					})

					It("does not check for versions of the parent resource type", func() {
						Expect(fakeDBPipeline.ResourceTypeCallCount()).To(Equal(1))
					})
				})
			})

			Context("when the resource type overrides a base resource type", func() {
				BeforeEach(func() {
					otherResourceType := fakeResourceType

					fakeResourceType = new(dbfakes.FakeResourceType)
					fakeResourceType.IDReturns(40)
					fakeResourceType.NameReturns("registry-image")
					fakeResourceType.TypeReturns("registry-image")
					fakeResourceType.SourceReturns(atc.Source{"custom": "((source-params))"})
					fakeResourceType.SetResourceConfigReturns(fakeResourceConfig, nil)

					fakeDBPipeline.ResourceTypesReturns([]db.ResourceType{
						fakeResourceType,
						otherResourceType,
					}, nil)
					fakeDBPipeline.ResourceTypeReturns(fakeResourceType, true, nil)
				})

				It("constructs the resource of the correct type", func() {
					Expect(fakeResourceType.SetResourceConfigCallCount()).To(Equal(1))
					_, resourceSource, resourceTypes := fakeResourceType.SetResourceConfigArgsForCall(0)
					Expect(resourceSource).To(Equal(atc.Source{"custom": "some-secret-sauce"}))
					Expect(resourceTypes).To(Equal(creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{
						versionedResourceType,
					})))

					_, owner, containerSpec, workerSpec, _ := fakePool.FindOrChooseWorkerForContainerArgsForCall(0)
					Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
					Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
						ResourceType: "registry-image",
					}))
					Expect(containerSpec.TeamID).To(Equal(123))
					Expect(workerSpec).To(Equal(worker.WorkerSpec{
						ResourceType:  "registry-image",
						ResourceTypes: creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{versionedResourceType}),
						TeamID:        123,
					}))

					Expect(fakeWorker.FindOrCreateContainerCallCount()).To(Equal(1))
					_, _, _, owner, metadata, containerSpec, resourceTypes := fakeWorker.FindOrCreateContainerArgsForCall(0)
					Expect(owner).To(Equal(db.NewResourceConfigCheckSessionContainerOwner(fakeResourceConfig, ContainerExpiries)))
					Expect(metadata).To(Equal(db.ContainerMetadata{
						Type: db.ContainerTypeCheck,
					}))
					Expect(containerSpec.ImageSpec).To(Equal(worker.ImageSpec{
						ResourceType: "registry-image",
					}))
					Expect(containerSpec.TeamID).To(Equal(123))
					Expect(resourceTypes).To(Equal(creds.NewVersionedResourceTypes(variables, atc.VersionedResourceTypes{
						versionedResourceType,
					})))
				})
			})

			It("grabs an immediate resource checking lock before checking, breaks lock after done", func() {
				Expect(fakeResourceConfig.AcquireResourceCheckingLockCallCount()).To(Equal(1))
				Expect(fakeResourceConfig.UpdateLastCheckedCallCount()).To(Equal(1))

				leaseInterval, immediate := fakeResourceConfig.UpdateLastCheckedArgsForCall(0)
				Expect(leaseInterval).To(Equal(interval))
				Expect(immediate).To(BeTrue())

				Eventually(fakeLock.ReleaseCallCount()).Should(Equal(1))
			})

			Context("when setting the resource config on the resource type fails", func() {
				BeforeEach(func() {
					fakeResourceType.SetResourceConfigReturns(nil, errors.New("catastrophe"))
				})

				It("sets the check error and returns the error", func() {
					Expect(runErr).To(HaveOccurred())
					Expect(fakeResourceType.SetCheckSetupErrorCallCount()).To(Equal(1))

					chkErr := fakeResourceType.SetCheckSetupErrorArgsForCall(0)
					Expect(chkErr).To(MatchError("catastrophe"))
				})
			})

			Context("when creating the container fails", func() {
				BeforeEach(func() {
					fakeWorker.FindOrCreateContainerReturns(nil, errors.New("catastrophe"))
				})

				It("sets the check error and returns the error", func() {
					Expect(runErr).To(HaveOccurred())
					Expect(fakeResourceConfig.SetCheckErrorCallCount()).To(Equal(1))

					chkErr := fakeResourceConfig.SetCheckErrorArgsForCall(0)
					Expect(chkErr).To(MatchError("catastrophe"))
				})
			})

			Context("when finding or choosing the worker fails", func() {
				BeforeEach(func() {
					fakePool.FindOrChooseWorkerForContainerReturns(nil, errors.New("catastrophe"))
				})

				It("sets the check error and returns the error", func() {
					Expect(runErr).To(HaveOccurred())
					Expect(fakeResourceConfig.SetCheckErrorCallCount()).To(Equal(1))

					chkErr := fakeResourceConfig.SetCheckErrorArgsForCall(0)
					Expect(chkErr).To(MatchError("catastrophe"))
				})
			})

			Context("when there is no current version", func() {
				BeforeEach(func() {
					fakeResourceType.VersionReturns(nil, nil)
				})

				It("checks from nil", func() {
					_, _, _, version := fakeResource.CheckArgsForCall(0)
					Expect(version).To(BeEmpty())
				})
			})

			Context("when there is a current version", func() {
				BeforeEach(func() {
					fakeResourceVersion1 := new(dbfakes.FakeResourceVersion)
					fakeResourceVersion1.IDReturns(1)
					fakeResourceVersion1.VersionReturns(db.Version{"version": "1"})
					fakeResourceVersion1.SpaceReturns(atc.Space("space1"))

					fakeResourceVersion2 := new(dbfakes.FakeResourceVersion)
					fakeResourceVersion2.IDReturns(2)
					fakeResourceVersion2.VersionReturns(db.Version{"version": "2"})
					fakeResourceVersion2.SpaceReturns(atc.Space("space2"))

					fakeResourceConfig.LatestVersionsReturns([]db.ResourceVersion{fakeResourceVersion1, fakeResourceVersion2}, nil)
				})

				It("checks with it", func() {
					Expect(fakeResource.CheckCallCount()).To(Equal(1))
					_, checkHandler, _, version := fakeResource.CheckArgsForCall(0)
					var expectedCheckHandler CheckEventHandler
					Expect(checkHandler).To(BeAssignableToTypeOf(&expectedCheckHandler))
					Expect(version).To(Equal(map[atc.Space]atc.Version{
						atc.Space("space1"): atc.Version{"version": "1"},
						atc.Space("space2"): atc.Version{"version": "2"},
					}))
				})
			})

			Context("when checking fails", func() {
				disaster := errors.New("nope")

				BeforeEach(func() {
					fakeResource.CheckReturns(disaster)
				})

				It("exits with the failure", func() {
					Expect(runErr).To(HaveOccurred())
					Expect(runErr).To(Equal(disaster))
				})
			})

			Context("when the lock is not immediately available", func() {
				BeforeEach(func() {
					results := make(chan bool, 4)
					results <- false
					results <- false
					results <- true
					results <- true
					close(results)

					fakeResourceConfig.AcquireResourceCheckingLockStub = func(logger lager.Logger) (lock.Lock, bool, error) {
						if <-results {
							return fakeLock, true, nil
						} else {
							// allow the sleep to continue
							go fakeClock.WaitForWatcherAndIncrement(time.Second)
							return nil, false, nil
						}
					}
				})

				It("retries every second until it is", func() {
					Expect(fakeResourceConfig.AcquireResourceCheckingLockCallCount()).To(Equal(3))
					Expect(fakeLock.ReleaseCallCount()).To(Equal(1))
				})
			})

			Context("when last checked interval is not immediately updated", func() {
				BeforeEach(func() {
					results := make(chan bool, 4)
					results <- false
					results <- false
					results <- true
					results <- true
					close(results)

					fakeResourceConfig.UpdateLastCheckedStub = func(interval time.Duration, immediate bool) (bool, error) {
						if <-results {
							return true, nil
						} else {
							// allow the sleep to continue
							go fakeClock.WaitForWatcherAndIncrement(time.Second)
							return false, nil
						}
					}
				})

				It("retries every second until it is", func() {
					Expect(fakeResourceConfig.AcquireResourceCheckingLockCallCount()).To(Equal(3))
					Expect(fakeResourceConfig.UpdateLastCheckedCallCount()).To(Equal(3))

					leaseInterval, immediate := fakeResourceConfig.UpdateLastCheckedArgsForCall(0)
					Expect(leaseInterval).To(Equal(interval))
					Expect(immediate).To(BeTrue())

					leaseInterval, immediate = fakeResourceConfig.UpdateLastCheckedArgsForCall(1)
					Expect(leaseInterval).To(Equal(interval))
					Expect(immediate).To(BeTrue())

					leaseInterval, immediate = fakeResourceConfig.UpdateLastCheckedArgsForCall(2)
					Expect(leaseInterval).To(Equal(interval))
					Expect(immediate).To(BeTrue())

					Expect(fakeLock.ReleaseCallCount()).To(Equal(3))
				})
			})

			It("clears the resource's check error", func() {
				Expect(fakeResourceConfig.SetCheckErrorCallCount()).To(Equal(1))

				err := fakeResourceConfig.SetCheckErrorArgsForCall(0)
				Expect(err).To(BeNil())
			})

			Context("when the pipeline is paused", func() {
				BeforeEach(func() {
					fakeDBPipeline.CheckPausedReturns(true, nil)
				})

				It("does not check", func() {
					Expect(fakeResource.CheckCallCount()).To(BeZero())
				})

				It("does not return an error", func() {
					Expect(runErr).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("ScanFromVersion", func() {
		var (
			fakeResource *rfakes.FakeResource
			fromVersion  map[atc.Space]atc.Version

			scanErr error
		)

		BeforeEach(func() {
			fakeWorker.NameReturns("some-worker")
			fakePool.FindOrChooseWorkerForContainerReturns(fakeWorker, nil)

			fakeContainer.HandleReturns("some-handle")
			fakeWorker.FindOrCreateContainerReturns(fakeContainer, nil)

			fakeResource = new(rfakes.FakeResource)
			fakeResourceFactory.NewResourceForContainerReturns(fakeResource, nil)
			fromVersion = nil
		})

		JustBeforeEach(func() {
			scanErr = scanner.ScanFromVersion(lagertest.NewTestLogger("test"), "some-resource-type", fromVersion)
		})

		Context("if the lock can be acquired", func() {
			BeforeEach(func() {
				fakeResourceConfig.AcquireResourceCheckingLockReturns(fakeLock, true, nil)
				fakeResourceConfig.UpdateLastCheckedReturns(true, nil)
			})

			Context("when fromVersion is nil", func() {
				It("checks from nil", func() {
					_, _, _, version := fakeResource.CheckArgsForCall(0)
					Expect(version).To(BeEmpty())
				})
			})

			Context("when fromVersion is specified", func() {
				BeforeEach(func() {
					fromVersion = map[atc.Space]atc.Version{
						atc.Space("space"): atc.Version{
							"version": "1",
						},
					}
				})

				It("checks from it", func() {
					_, _, _, version := fakeResource.CheckArgsForCall(0)
					Expect(version).To(Equal(fromVersion))
				})

				Context("when there are no latest versions", func() {
					It("checks from the fromVersion", func() {
						_, _, _, version := fakeResource.CheckArgsForCall(0)
						Expect(version).To(Equal(fromVersion))
					})
				})

				Context("when there are latest versions", func() {
					BeforeEach(func() {
						fakeResourceVersion1 := new(dbfakes.FakeResourceVersion)
						fakeResourceVersion1.IDReturns(1)
						fakeResourceVersion1.VersionReturns(db.Version{"version": "1"})
						fakeResourceVersion1.SpaceReturns(atc.Space("space1"))

						fakeResourceVersion2 := new(dbfakes.FakeResourceVersion)
						fakeResourceVersion2.IDReturns(2)
						fakeResourceVersion2.VersionReturns(db.Version{"version": "2"})
						fakeResourceVersion2.SpaceReturns(atc.Space("space2"))

						fakeResourceConfig.LatestVersionsReturns([]db.ResourceVersion{fakeResourceVersion1, fakeResourceVersion2}, nil)
					})

					Context("when fromVersion has the same space", func() {
						BeforeEach(func() {
							fromVersion = map[atc.Space]atc.Version{
								atc.Space("space1"): atc.Version{
									"version": "2",
								},
							}
						})

						It("checks from the correct version", func() {
							_, _, _, version := fakeResource.CheckArgsForCall(0)
							Expect(version).To(Equal(map[atc.Space]atc.Version{
								atc.Space("space1"): atc.Version{"version": "2"},
								atc.Space("space2"): atc.Version{"version": "2"},
							}))
						})
					})

					Context("when fromVersion does not have the same space", func() {
						It("checks from the correct version", func() {
							_, _, _, version := fakeResource.CheckArgsForCall(0)
							Expect(version).To(Equal(map[atc.Space]atc.Version{
								atc.Space("space"):  atc.Version{"version": "1"},
								atc.Space("space1"): atc.Version{"version": "1"},
								atc.Space("space2"): atc.Version{"version": "2"},
							}))
						})
					})
				})
			})

			Context("when checking fails with ErrResourceScriptFailed", func() {
				scriptFail := atc.ErrResourceScriptFailed{}

				BeforeEach(func() {
					fakeResource.CheckReturns(scriptFail)
				})

				It("returns the error", func() {
					Expect(scanErr).To(Equal(scriptFail))
				})
			})

			Context("when the resource is not in the database", func() {
				BeforeEach(func() {
					fakeDBPipeline.ResourceTypeReturns(nil, false, nil)
				})

				It("returns an error", func() {
					Expect(scanErr).To(HaveOccurred())
					Expect(scanErr.Error()).To(ContainSubstring("resource type not found: some-resource-type"))
				})
			})
		})
	})
})
