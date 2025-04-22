import React, { useState } from 'react';
import AuthForm from '../components/AuthForm';
import AuthLayout from '../layouts/AuthLayout';
import { resetPassword } from '../utils/api';

const ForgotPassword = () => {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [emailSent, setEmailSent] = useState(false);

  const handleSubmit = async (e) => {
    e?.preventDefault();
    
    if (!email) {
      alert('Please enter your email address');
      return;
    }

    setIsLoading(true);
    try {
      await resetPassword({ email });
      setEmailSent(true);
    } catch (error) {
      console.error('Password reset error:', error);
      alert(error.response?.data?.error || 'Failed to send reset email. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (emailSent) {
    return (
      <AuthLayout>
        <div className="auth-container">
          <h2>Check Your Email</h2>
          <p>
            We've sent a password reset link to: <strong>{email}</strong>
          </p>
          <p>
            Please check your email and click the link to reset your password.
            The link will expire in 15 minutes.
          </p>
          <p>
            <a href="/login">Back to Login</a>
          </p>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout>
      <AuthForm
        title="Forgot Password"
        fields={[
          {
            label: 'Email',
            type: 'email',
            value: email,
            onChange: (e) => setEmail(e.target.value),
            placeholder: 'Enter your email address',
            disabled: isLoading
          },
        ]}
        onSubmit={handleSubmit}
        submitButtonText={isLoading ? "Sending Reset Link..." : "Send Reset Link"}
        isLoading={isLoading}
        footer={
          <div>
            <a href="/login">Back to Login</a>
          </div>
        }
      />
    </AuthLayout>
  );
};

export default ForgotPassword;
