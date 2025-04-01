import React from 'react';

const InputField = ({ label, type, value, onChange, placeholder }) => {
    const id = label.toLowerCase().replace(/[^a-z0-9]/g, '-');
    
    return (
        <div className="input-container">
            <label htmlFor={id}>{label}</label>
            <input 
                id={id}
                type={type} 
                value={value} 
                onChange={onChange} 
                placeholder={placeholder} 
                required 
            />
        </div>
    );
};

export default InputField;
