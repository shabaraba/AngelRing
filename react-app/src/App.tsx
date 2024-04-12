import React, { useState, useEffect } from 'react';
import Dropzone from 'react-dropzone';
import Modal from 'react-modal';

Modal.setAppElement('#root');

const App: React.FC = () => {
  const [files, setFiles] = useState<string[]>([]);
  const [modalIsOpen, setModalIsOpen] = useState<boolean>(false);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [uploading, setUploading] = useState<boolean>(false);

  useEffect(() => {
    setFiles(['/static/images/sample1.png','/static/images/sample2.png'])
  }, []);

  const openModal = (image: string) => {
    setSelectedImage(image);
    setModalIsOpen(true);
  };

  const closeModal = () => {
    setModalIsOpen(false);
  };

  const handleUpload = async () => {
    setUploading(true);
    try {
      const formData = new FormData();
      files.forEach((file) => {
        formData.append('files', file);
      });
      const response = await fetch('http://your-api-endpoint/upload', {
        method: 'POST',
        body: formData,
      });
      if (response.ok) {
        console.log('Upload successful');
        setFiles([]);
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
      <Dropzone onDrop={(acceptedFiles) => setFiles([...files, ...acceptedFiles.map(file => URL.createObjectURL(file))])}>
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
        {files.map((file, index) => (
          <img
            key={index}
            src={file}
            alt={`Uploaded file ${index}`}
            onClick={() => openModal(file)}
            style={{ width: '100px', height: '100px', margin: '10px', cursor: 'pointer' }}
          />
        ))}
      </div>
      <button onClick={handleUpload} disabled={files.length === 0 || uploading}>
        {uploading ? 'Uploading...' : 'Upload'}
      </button>
      <Modal isOpen={modalIsOpen} onRequestClose={closeModal}>
        {selectedImage && <img src={selectedImage} alt="Full size" style={{ maxWidth: '100%' }} />}
      </Modal>
    </div>
  );
};

export default App;

