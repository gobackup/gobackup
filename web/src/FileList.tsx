import { Button } from 'antd';
import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { PageTitle } from './components';

const FileList: FC<{}> = () => {
  let { model = '' } = useParams();

  const [files, setFiles] = useState<any[]>([]);
  const [parent, setParent] = useState('/');

  useEffect(() => {
    let query = new URLSearchParams({
      model,
      parent,
    });

    fetch(`/api/list?` + query.toString())
      .then((res) => res.json())
      .then((data) => {
        setFiles(data.files);
      });
  }, [model]);

  const FileItem = ({
    file,
    type = 'file',
  }: {
    file: any;
    type?: 'file' | 'folder';
  }) => {
    return (
      <div className="flex items-center justify-between space-x-2 py-2 px-6 hover:bg-gray-50">
        <div>{file.filename}</div>
        {type === 'file' && (
          <>
            <div className="flex items-center text-sm space-x-2">
              <div>{file.size}</div>
              <div>{file.last_modified}</div>
            </div>
            <div>
              <Button size="small">Download</Button>
            </div>
          </>
        )}
      </div>
    );
  };

  return (
    <div>
      <PageTitle title={`Browser: ${model}`} backTo={`/`} />
      <div className="rounded border shadow-sm mt-4 divide-y divide-gray-100">
        <FileItem key={0} file={{ filename: parent }} type="folder" />
        {files.map((file, i) => (
          <FileItem key={i} file={file} />
        ))}
      </div>
    </div>
  );
};

export default FileList;
