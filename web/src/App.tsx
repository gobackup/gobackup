import { Button, notification, Popconfirm, Skeleton } from 'antd';
import { useEffect, useState } from 'react';
import { LazyLog, ScrollFollow } from 'react-lazylog';
import { Link } from 'react-router-dom';
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
      <div className="model-list-item">
        <div className="text-base">
          <div className="text-base font-medium uppercase">{modelKey}</div>
          {scheduleEnable && (
            <div className="text-green text-sm">{model.schedule_info}</div>
          )}
          {model.description && (
            <div className="text-gray-400 truncate text-xs my-1">
              {model.description}
            </div>
          )}
        </div>
        <div className="flex items-center space-x-1">
          <Link to={`/browser/${modelKey}`}>
            <Button size="small">
              <Icon name="folders" />
            </Button>
          </Link>

          <Popconfirm
            title="Perform Backup"
            description="Are you sure to perform backup now?"
            onConfirm={() => performBackup(modelKey)}
          >
            <Button size="small" title="Perform backup now!">
              <Icon name="play" mode="fill" />
            </Button>
          </Popconfirm>
        </div>
      </div>
    );
  };

  return (
    <div className="model-list-wrapper">
      <div className="model-list-header">
        <div className="flex items-center space-x-2">
          <Icon name="stack" />
          <div className="text-text text-base">Models</div>
        </div>
      </div>
      <div className="model-list-scrollview">
        {loading && (
          <div className="p-4">
            <Skeleton active />
          </div>
        )}
        {!loading && (
          <>
            {Object.keys(models).map((key: string, idx: number) => (
              <ModelItem modelKey={key} key={idx} />
            ))}
          </>
        )}
      </div>
    </div>
  );
};

const App = () => {
  return (
    <div className="flex flex-col relative md:flex-row gap-4">
      <ModelList />
      <LogView />
    </div>
  );
};

export default App;
