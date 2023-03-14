import { Button, notification } from 'antd';
import 'antd/dist/reset.css';
import { useEffect, useState } from 'react';
import Icon from './icon';

const API_URL = '/api';

const ModelList = ({}) => {
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState([]);

  useEffect(() => {
    reloadModels();
  }, []);

  const performBackup = (model: string) => {
    fetch(`${API_URL}/perform`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ model }),
    })
      .then((res) => res.json())
      .then((data) => {
        notification.success({
          message: 'Backup',
          description: `Backup for ${model} performed successfully.`,
        });
      })
      .catch((data) => {
        notification.error({
          message: 'Backup Failed',
          description: data.message,
        });
      });
  };

  const reloadModels = () => {
    setLoading(true);
    fetch(`${API_URL}/config`)
      .then((res) => res.json())
      .then((config) => {
        setModels(config.models);
        setLoading(false);
      });
  };

  return (
    <div className="rounded border border-gray-200">
      <div className="text-lg text-gray-600 p-2 px-4 bg-gray-100 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="text-text">Models</div>
          <Button
            icon={<Icon name="refresh" />}
            size="small"
            onClick={reloadModels}
          >
            Refresh
          </Button>
        </div>
      </div>
      {loading && <>Loading...</>}
      {!loading && (
        <ul className="divide-y">
          {models.map((model: string) => (
            <li
              key={model}
              className="flex items-center justify-between py-2 px-4 hover:bg-gray-50"
            >
              <div className="text-base">{model}</div>
              <Button
                icon={<Icon name="play" mode="fill" />}
                size="small"
                onClick={() => performBackup(model)}
              >
                Backup
              </Button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

const App = () => {
  return (
    <div className="py-6">
      <img
        src="https://user-images.githubusercontent.com/5518/205909959-12b92929-4ac5-4bb5-9111-6f9a3ed76cf6.png"
        className="h-24 mx-auto"
      />
      <div className="container max-w-3xl mx-auto mt-10 rounded bg-white shadow-sm p-6 border border-gray-200">
        <ModelList />
      </div>
      <div className="footer">
        <div className="copyright">GoBackup Powered.</div>
        <div className="links">
          <a href="https://gobackup.github.io" target="_blank">
            <Icon name="home-smile" mode="fill" />
          </a>
          <a href="https://github.com/gobackup/gobackup" target="_blank">
            <Icon name="github" mode="fill" />
          </a>
        </div>
      </div>
    </div>
  );
};

export default App;
