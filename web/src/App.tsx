import { useEffect, useState } from 'react';

const API_URL = '/api';

const ModelList = ({}) => {
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState([]);

  const performBackup = (model: string) => {
    fetch(`${API_URL}/perform`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ model }),
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

  useEffect(() => {
    reloadModels();
  }, []);

  return (
    <div className="rounded border border-gray-200">
      <div className="text-lg text-gray-600 p-2 px-4 bg-gray-100 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>Models</div>
          <button className="btn" onClick={reloadModels}>
            Reload
          </button>
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
              <button className="btn" onClick={() => performBackup(model)}>
                Backup Now
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

const App = () => {
  return (
    <div className="container max-w-3xl mx-auto my-20 rounded bg-white shadow p-6 border border-gray-100">
      <div className="border-b border-gray-200 pb-3 mb-6">
        <a href="/" className="text-2xl font-semibold text-blue-700">
          GoBackup
        </a>
      </div>
      <ModelList />
    </div>
  );
};

export default App;
