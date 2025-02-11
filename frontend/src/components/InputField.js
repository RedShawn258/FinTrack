import React from 'react';

const InputField = ({ label, type, value, onChange, placeholder }) => {
    return (
        <div className="input-container">
            <label>{label}</label>
            <input type={type} value={value} onChange={onChange} placeholder={placeholder} required />
        </div>
    );
};

export default InputField;
