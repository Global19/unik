package state

import (
	"encoding/json"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type State interface {
	GetImages() map[string]*types.Image
	GetInstances() map[string]*types.Instance
	GetVolumes() map[string]*types.Volume
	ModifyImages(modify func(images map[string]*types.Image) error) error
	ModifyInstances(modify func(instances map[string]*types.Instance) error) error
	ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error
	Save() error
	Load() error
}

type memoryState struct {
	imagesLock    sync.RWMutex
	instancesLock sync.RWMutex
	volumesLock   sync.RWMutex
	saveLock      sync.Mutex
	saveFile      string
	Images        map[string]*types.Image    `json:"Images"`
	Instances     map[string]*types.Instance `json:"Instances"`
	Volumes       map[string]*types.Volume   `json:"Volumes"`
}

func NewMemoryState(saveFile string) *memoryState {
	return &memoryState{
		saveFile:      saveFile,
		Images:        make(map[string]*types.Image),
		Instances:     make(map[string]*types.Instance),
		Volumes:       make(map[string]*types.Volume),
	}
}

func (s *memoryState) GetImages() map[string]*types.Image {
	s.imagesLock.RLock()
	defer s.imagesLock.RUnlock()
	imagesCopy := make(map[string]*types.Image)
	for id, image := range s.Images {
		imageCopy := &types.Image{
			Id:             image.Id,
			Name:           image.Name,
			DeviceMappings: image.DeviceMappings,
			SizeMb:         image.SizeMb,
			Infrastructure: image.Infrastructure,
			Created:        image.Created,
		}
		imagesCopy[id] = imageCopy
	}
	return imagesCopy
}

func (s *memoryState) GetInstances() map[string]*types.Instance {
	s.instancesLock.RLock()
	defer s.instancesLock.RUnlock()
	instancesCopy := make(map[string]*types.Instance)
	for id, instance := range s.Instances {
		instanceCopy := &types.Instance{
			Id:             instance.Id,
			ImageId:        instance.ImageId,
			Infrastructure: instance.Infrastructure,
			Name:           instance.Name,
			State:          instance.State,
			Created:          instance.Created,
		}
		instancesCopy[id] = instanceCopy
	}
	return instancesCopy
}

func (s *memoryState) GetVolumes() map[string]*types.Volume {
	s.volumesLock.RLock()
	defer s.volumesLock.RUnlock()
	volumesCopy := make(map[string]*types.Volume)
	for id, volume := range s.Volumes {
		volumeCopy := &types.Volume{
			Id:             volume.Id,
			Name:           volume.Name,
			SizeMb:         volume.SizeMb,
			Attachment:     volume.Attachment,
			Infrastructure: volume.Infrastructure,
			Created: volume.Created,
		}
		volumesCopy[id] = volumeCopy
	}
	return volumesCopy
}

func (s *memoryState) ModifyImages(modify func(images map[string]*types.Image) error) error {
	s.imagesLock.Lock()
	defer s.imagesLock.Unlock()
	return modify(s.Images)
}

func (s *memoryState) ModifyInstances(modify func(instances map[string]*types.Instance) error) error {
	s.instancesLock.Lock()
	defer s.instancesLock.Unlock()
	return modify(s.Instances)
}

func (s *memoryState) ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error {
	s.volumesLock.Lock()
	defer s.volumesLock.Unlock()
	return modify(s.Volumes)
}

func (s *memoryState) Save() error {
	s.saveLock.Lock()
	defer s.saveLock.Unlock()
	data, err := json.Marshal(s)
	if err != nil {
		return lxerrors.New("failed to marshal memory state to json", err)
	}
	os.MkdirAll(filepath.Dir(s.saveFile), 0751)
	err = ioutil.WriteFile(s.saveFile, data, 0644)
	if err != nil {
		return lxerrors.New("writing save file "+s.saveFile, err)
	}
	return nil
}

func (s *memoryState) Load() error {
	data, err := ioutil.ReadFile(s.saveFile)
	if err != nil {
		return lxerrors.New("error reading save file "+s.saveFile, err)
	}
	var newS memoryState
	err = json.Unmarshal(data, &newS)
	if err != nil {
		return lxerrors.New("failed to unmarshal data "+string(data)+" to memory state", err)
	}
	newS.saveFile = s.saveFile
	*s = newS
	return nil
}
