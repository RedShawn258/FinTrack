import React from 'react';
import './AuthLayout.css';

const AuthLayout = ({ children }) => {
  const backgroundStyle = {
    backgroundImage: "url('/assets/sample.jpg')",
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
    height: '100vh',
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center'
  };

  return <div style={backgroundStyle}>{children}</div>;
};

export default AuthLayout; 