import { ConfigProvider } from 'antd';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

import 'remixicon/fonts/remixicon.css';
import './style.scss';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#EB5424',
          colorText: '#313638',
          colorSuccess: '#4BAB4E',
          colorError: '#EB5424',
          colorInfo: '#2454BB',
          colorBgBase: '#E8E9EB',
          borderRadius: 4,
        },
      }}
    >
      <App />
    </ConfigProvider>
  </React.StrictMode>
);
