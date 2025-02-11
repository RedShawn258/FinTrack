import React from 'react';
import InputField from './InputField';
import Button from './Button';

const AuthForm = ({ title, fields, onSubmit, footer }) => {
    return (
        <div className="auth-container">
            <h2>{title}</h2>
            {fields.map(({ label, type, value, onChange }, index) => (
                <InputField key={index} label={label} type={type} value={value} onChange={onChange} />
            ))}
            <Button text={title} onClick={onSubmit} />
            {footer}
        </div>
    );
};

export default AuthForm;
