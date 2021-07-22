import React from 'react';

type TProps = {
    className?: string
};
export const DownArrowIcon = ({
    className = '',
}: TProps) => <svg
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 13l-5 5m0 0l-5-5m5 5V6" />
</svg>