import React, { useState, useEffect } from 'react';
import Dropzone from 'react-dropzone';
import Modal from 'react-modal';

Modal.setAppElement('#root');

type Pathes = {
  path: string; 
  thumbnailPath: string;
}

const App: React.FC = () => {
  const [uploadingFiles, setUploadingFiles] = useState<File[]>([]);
  const [filePathes, setFilePathes] = useState<Array<Pathes>>([]);
  const [modalIsOpen, setModalIsOpen] = useState<boolean>(false);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [uploading, setUploading] = useState<boolean>(false);

  useEffect(() => {
    fetch('/api/images')
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        return response.json();
      })
      .then(response => {
        const data = JSON.parse(response.data);
        console.log(data);
        setFilePathes(data.map((d: Pathes) => {return {"path": d.path, "thumbnailPath": d.thumbnailPath}}));
      })
      .catch(error => {
        console.log(error);
        // setError(error.message);
        // setLoading(false);
      });
    setUploadingFiles([])
  }, [uploading]);

  const openModal = (path: string) => {
    setSelectedImage(path);
    setModalIsOpen(true);
  };

  const closeModal = () => {
    setModalIsOpen(false);
  };

  const handleUpload = async () => {
    setUploading(true);
    try {
      const formData = new FormData();
      uploadingFiles.forEach((file) => {
        formData.append('files', file);
      });
      const response = await fetch('/api/images', {
        headers: {
        },
        method: 'POST',
        body: formData,
      });
      if (response.ok) {
        console.log('Upload successful');
        setUploadingFiles([]);
      } else {
        console.error('Upload failed');
      }
    } catch (error) {
      console.error('Error uploading files:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <Dropzone onDrop={(acceptedFiles) => setUploadingFiles([...uploadingFiles, ...acceptedFiles])}>
        {({ getRootProps, getInputProps }) => (
          <section>
            <div {...getRootProps()} style={{ border: '1px solid black', padding: '20px', textAlign: 'center', cursor: 'pointer' }}>
              <input {...getInputProps()} />
              <p>Drag & drop some files here, or click to select files</p>
            </div>
          </section>
        )}
      </Dropzone>
      <div style={{ display: 'flex', flexWrap: 'wrap' }}>
        {uploadingFiles.map((file, index) => (
          <img
            key={index}
            src={URL.createObjectURL(file)}
            alt={`Uploaded file ${index}`}
            onClick={() => openModal(URL.createObjectURL(file))}
            style={{ width: '100px', height: '100px', margin: '10px', cursor: 'pointer' }}
          />
        ))}
      </div>
      <button onClick={handleUpload} disabled={uploadingFiles.length === 0 || uploading}>
        {uploading ? 'Uploading...' : 'Upload'}
      </button>
      <div style={{ display: 'flex', flexWrap: 'wrap' }}>
        {filePathes.map((filePath, index) => (
          <img
            key={index}
            src={filePath.thumbnailPath}
            alt={`Uploaded file ${index}`}
            onClick={() => openModal(filePath.path)}
            style={{ width: '100px', height: '100px', margin: '10px', cursor: 'pointer' }}
          />
        ))}
      </div>
      <Modal isOpen={modalIsOpen} onRequestClose={closeModal}>
        {selectedImage && <img src={selectedImage} alt="Full size" style={{ maxWidth: '100%' }} />}
      </Modal>
    </div>
  );
};

export default App;

