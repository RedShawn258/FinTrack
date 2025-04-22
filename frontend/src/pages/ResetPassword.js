import React, { useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { resetPasswordWithToken } from '../utils/api';
import AuthForm from '../components/AuthForm';
import AuthLayout from '../layouts/AuthLayout';

const ResetPassword = () => {
  const [params] = useSearchParams();
  const token = params.get('token');
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);

  const [form, setForm] = useState({
    newPassword: '',
    confirmPassword: ''
  });

  const handleSubmit = async (e) => {
    e?.preventDefault();
    
    const { newPassword, confirmPassword } = form;

    if (!newPassword || !confirmPassword) {
      alert('Please fill out both password fields');
      return;
    }

    if (newPassword !== confirmPassword) {
      alert('Passwords do not match');
      return;
    }

    if (!token) {
      alert('Invalid reset token. Please request a new password reset.');
      navigate('/forgot-password');
      return;
    }

    setIsLoading(true);
    try {
      await resetPasswordWithToken({ token, password: newPassword });
      alert('Password has been reset successfully! Please login with your new password.');
      navigate('/login');
    } catch (error) {
      console.error('Password reset error:', error);
      alert(error.response?.data?.error || 'Failed to reset password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (!token) {
    return (
      <AuthLayout>
        <div className="auth-container">
          <h2>Invalid Reset Link</h2>
          <p>This password reset link is invalid or has expired.</p>
          <p>
            <a href="/forgot-password">Request a new password reset</a>
          </p>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout>
      <AuthForm
        title="Reset Your Password"
        fields={[
          {
            label: 'New Password',
            type: 'password',
            value: form.newPassword,
            onChange: (e) => setForm({ ...form, newPassword: e.target.value }),
            placeholder: 'Enter your new password'
          },
          {
            label: 'Confirm Password',
            type: 'password',
            value: form.confirmPassword,
            onChange: (e) => setForm({ ...form, confirmPassword: e.target.value }),
            placeholder: 'Confirm your new password'
          }
        ]}
        onSubmit={handleSubmit}
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

export default ResetPassword;
