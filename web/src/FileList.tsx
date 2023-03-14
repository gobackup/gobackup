import { Button } from 'antd';
import { filesize } from 'filesize';
import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { PageTitle } from './components';

import Icon from './icon';

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
    const downloadURL =
      `/api/download?` +
      new URLSearchParams({
        model,
        path: file.filename,
      }).toString();

    const fsize = filesize(file.size || 0, { base: 2 }).toString();

    return (
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-2 py-2 px-6 hover:bg-gray-50">
        <div>{file.filename}</div>
        {type === 'file' && (
          <>
            <div className="flex items-center justify-between text-sm space-x-4 text-gray-400">
              <div>{fsize}</div>
              <div>{file.last_modified}</div>
              <div>
                <Button size="small" title="Download backup file.">
                  <a href={downloadURL}>
                    <Icon name="download-cloud" mode="fill" />
                  </a>
                </Button>
              </div>
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
