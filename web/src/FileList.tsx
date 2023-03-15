import { Button, Empty, Skeleton } from 'antd';
import { filesize } from 'filesize';
import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { PageTitle } from './components';

import Icon from './icon';

const FileList: FC<{}> = () => {
  let { model = '' } = useParams();

  const [loading, setLoading] = useState(true);
  const [files, setFiles] = useState<any[]>([]);
  const [parent, setParent] = useState('/');

  const Time = ({ value }: { value: string }) => {
    if (!value) return <></>;
    return <span title={value}>{new Date(value).toLocaleString()}</span>;
  };

  const reloadList = () => {
    setLoading(true);
    let query = new URLSearchParams({
      model,
      parent,
    });

    fetch(`/api/list?` + query.toString())
      .then((res) => res.json())
      .then((data) => {
        setFiles(data.files || []);
        setLoading(false);
      });
  };

  useEffect(() => {
    reloadList();
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
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-2 py-2 px-2 hover:bg-gray-50">
        <a
          className="flex items-center space-x-2 hover:text-blue"
          href={downloadURL}
        >
          <Icon name="folder-zip" />
          <div className="max-w-xl truncate">{file.filename}</div>
        </a>
        {type === 'file' && (
          <>
            <div className="flex items-center justify-between text-sm space-x-4 text-gray-400">
              <div>{fsize}</div>
              <div>
                <Time value={file.last_modified} />
              </div>
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
      <PageTitle
        title={
          <div className="flex lg:items-center flex-col lg:flex-row-reverse lg:gap-x-2">
            <div className="text-xs text-gray-600">Browser</div>
            <div className="uppercase text-base">{model}</div>
          </div>
        }
        backTo={`/`}
        extra={
          <>
            <Button size="small" onClick={reloadList} title="Refresh">
              <Icon name="refresh" loading={loading} />
            </Button>
          </>
        }
      />
      <div className="file-browser-container">
        {loading && <Skeleton active />}
        {!loading && (
          <>
            {files.length === 0 && <Empty className="pt-10" />}
            {files.map((file, i) => (
              <FileItem key={i} file={file} />
            ))}
          </>
        )}
      </div>
    </div>
  );
};

export default FileList;
