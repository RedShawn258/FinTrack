import React from 'react';
import InputField from './InputField';
import Button from './Button';
import AuthLayout from '../layouts/AuthLayout';
import './AuthForm.css';

const AuthForm = ({ title, fields, onSubmit, footer, isLoading = false }) => {
    return (
        <AuthLayout>
            <div className="auth-container">
                <h2>{title}</h2>
                {fields.map(({ label, type, value, onChange }, index) => (
                    <InputField key={index} label={label} type={type} value={value} onChange={onChange} />
                ))}
                <Button 
                    text={isLoading ? "Loading..." : title} 
                    onClick={onSubmit} 
                    disabled={isLoading} 
                />
                {footer}
            </div>
        </AuthLayout>
    );
};

export default AuthForm;
