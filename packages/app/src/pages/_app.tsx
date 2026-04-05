import '@clipboard/styles/globals.css';
import { HeadTemplate } from '../templates/HeadTemplate';
import type { AppProps } from 'next/app';
import { FC } from 'react';

const App: FC<AppProps> = ({ Component, pageProps }) => {
  return (
    <>
      <HeadTemplate basic={{ title: 'Clipboard' }} />
      <Component {...pageProps} />
    </>
  );
};

export default App;
