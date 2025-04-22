import React from 'react';
import './AuthForm.css';

const AuthForm = ({ 
    title, 
    fields, 
    onSubmit, 
    footer,
    isLoading,
    error,
    submitButtonText = "Submit"
}) => {
    const handleSubmit = (e) => {
        e.preventDefault();
        if (!isLoading) {
            onSubmit(e);
        }
    };

    return (
        <div className="auth-container">
            <h2>{title}</h2>
            {error && (
                <div className="error-message">
                    {error}
                </div>
            )}
            <form onSubmit={handleSubmit}>
                {fields.map((field, index) => (
                    <div key={index} className="form-group">
                        <label className="visually-hidden">{field.label}</label>
                        <input
                            type={field.type}
                            value={field.value}
                            onChange={field.onChange}
                            placeholder={field.label}
                            disabled={field.disabled || isLoading}
                            required
                        />
                    </div>
                ))}
                <button 
                    type="submit" 
                    className={`submit-button ${isLoading ? 'loading' : ''}`}
                    disabled={isLoading}
                >
                    {isLoading ? 'Please wait...' : submitButtonText}
                </button>
            </form>
            {footer && <div className="auth-links">{footer}</div>}
        </div>
    );
};

export default AuthForm;
