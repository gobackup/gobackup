import { Button, notification, Popconfirm } from 'antd';
import { useEffect, useState } from 'react';
import { LazyLog, ScrollFollow } from 'react-lazylog';
import Icon from './icon';

const API_URL = '/api';

/**
 * LazyLog
 *
 * https://mozilla-frontend-infra.github.io/react-lazylog/
 * https://github.com/mozilla-frontend-infra/react-lazylog
 */
const LogView = () => {
  return (
    <div className="log-wrapper">
      <ScrollFollow
        startFollowing
        render={({ follow, onScroll }) => (
          <LazyLog
            extraLines={1}
            enableSearch={true}
            url={`${API_URL}/log`}
            stream
            follow={follow}
            onScroll={onScroll}
          />
        )}
      />
    </div>
  );
};

const ModelList = ({}) => {
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState({});

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
      .then((data) => {
        setModels(data.models);
        setLoading(false);
      });
  };

  const ModelItem = ({ modelKey }: { modelKey: string }) => {
    const model = models[modelKey];
    const scheduleEnable = model.schedule?.enabled;

    return (
      <li
        key={modelKey}
        className="flex items-center justify-between py-2 px-4 hover:bg-gray-50"
      >
        <div className="divider-y text-base">
          <div className="text-base font-medium">{modelKey}</div>
          {scheduleEnable && (
            <div className="text-green text-sm">{model.schedule_info}</div>
          )}
        </div>
        <Popconfirm
          title="Perform Backup"
          description="Are you sure to perform backup now?"
          onConfirm={() => performBackup(modelKey)}
        >
          <Button size="small" title="Perform backup now!">
            <Icon name="play" mode="fill" />
          </Button>
        </Popconfirm>
      </li>
    );
  };

  return (
    <div className="rounded w-full xl:w-[500px] border border-gray-200">
      <div className="text-lg text-gray-600 p-2 px-4 bg-gray-100 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="text-text">Models</div>
          <Button size="small" onClick={reloadModels} title="Refresh">
            <Icon name="refresh" />
          </Button>
        </div>
      </div>
      {loading && <>Loading...</>}
      {!loading && (
        <ul className="divide-y">
          {Object.keys(models).map((key: string) => (
            <ModelItem modelKey={key} />
          ))}
        </ul>
      )}
    </div>
  );
};

const App = () => {
  return (
    <div className="py-6">
      <div className="container rounded bg-white shadow-sm p-6 border border-gray-200">
        <div className="flex flex-col xl:flex-row gap-4">
          <ModelList />
          <LogView />
        </div>
      </div>

      <div className="footer container">
        <div className="copyright flex items-center space-x-1">
          <img
            src="https://user-images.githubusercontent.com/5518/205909959-12b92929-4ac5-4bb5-9111-6f9a3ed76cf6.png"
            className="h-6 mx-auto"
          />
          <div>
            <a href="https://github.com/gobackup/gobackup" target="_blank">
              GoBackup
            </a>{' '}
            powered.
          </div>
        </div>
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
