import { StyleProvider } from '@ant-design/cssinjs';
import { ConfigProvider } from 'antd';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import App from './App';
import FileList from './FileList';

import 'remixicon/fonts/remixicon.css';
import Icon from './icon';
import './style.scss';

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
  },
  {
    path: `/browser/:model`,
    element: <FileList />,
  },
]);

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
          borderRadius: 4,
        },
      }}
    >
      <StyleProvider hashPriority="high">
        <React.StrictMode>
          <div className="py-6">
            <div className="container rounded bg-white shadow-sm p-6 border border-gray-200">
              <RouterProvider router={router} />
            </div>
            <div className="footer container">
              <div className="copyright flex items-center space-x-1">
                <img
                  src="https://user-images.githubusercontent.com/5518/205909959-12b92929-4ac5-4bb5-9111-6f9a3ed76cf6.png"
                  className="h-6 mx-auto"
                />
                <div>
                  <a
                    href="https://github.com/gobackup/gobackup"
                    target="_blank"
                  >
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
        </React.StrictMode>
      </StyleProvider>
    </ConfigProvider>
  </React.StrictMode>
);
