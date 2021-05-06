import React from 'react';
import './button.scss';
import classnames from 'classnames';

type TProps = {
    ghost?       : boolean;
    success?     : boolean;
    danger?      : boolean;
    largeIcon?   : boolean;
    small?       : boolean;
    outlined?    : boolean;
    hidden?      : boolean;
    bgVisible?   : boolean;
    label?       : string;
    hideOnMobile?: boolean;
    absolute?    : boolean;
    onClick?     : () => void;
    className?   : string;
    icon?        : JSX.Element | null
}
export const Button = ({
    ghost     = false,
    success   = false,
    danger    = false,
    largeIcon = false,
    small     = false,
    outlined  = false,
    hidden    = false,
    bgVisible = false,
    label,
    hideOnMobile = false,
    absolute     = false,
    onClick,
    className = '',
    children,
    icon = null,
}: React.PropsWithChildren<TProps>) => {

    const classes = classnames(
        'button-component',
        {
            '--ghost'         : ghost,
            '--success'       : success,
            '--danger'        : danger,
            '--large-icon'    : largeIcon,
            '--small'         : small,
            '--hide-on-mobile': hideOnMobile,
            '--outlined'      : outlined,
            '--hidden'        : hidden,
            '--bg-visible'    : bgVisible,
            '--absolute'      : absolute,
        },
        className
    )

    return <div
        className={classes}
        onClick={onClick}>
        {label && <span>{label}</span>}
        {children}
        {icon}
    </div>
}