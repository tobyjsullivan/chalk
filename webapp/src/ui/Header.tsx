import React from 'react';

import './Header.css';

interface PropsType {
  title: string,
}

const Header = ({title}: PropsType) => (
  <h1 className="Header-title">{title}</h1>
);

export default Header;
