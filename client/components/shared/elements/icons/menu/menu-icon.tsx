import React from 'react';

type TProps = {
    className?: string
};
export const MenuIcon = ({
    className = '',
}: TProps) => <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    className={className}
    viewBox="0 0 24 24" stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
</svg>;