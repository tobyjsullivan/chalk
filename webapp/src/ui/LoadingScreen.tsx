import React from 'react';

import Header from './Header';

interface PropsType {
  title: string,
}

const LoadingScreen = ({title}: PropsType) => (
  <div className="LoadingScreen">
    <Header title={title} />
    <p>Loading...</p>
  </div>
);

export default LoadingScreen;
