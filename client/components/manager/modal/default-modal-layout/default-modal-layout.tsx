import React from 'react';
import './default-modal-layout.scss';
import classnames from 'classnames';

type TCommonProps = {
    className?: string;
}

export const DefaultModalLayout = (props: React.PropsWithChildren<TCommonProps>) =>
    <div className={classnames('default-modal-layout', props.className)}>
        {props.children}
    </div>;

type THeaderProps = TCommonProps & {
    title?: string;
    subtitle?: string;
};
export const DefaultModalHeader = (props: React.PropsWithChildren<THeaderProps>) =>
    <div className={classnames('__header', props.className)}>
        <div>
            {props.title}
            {props.children}
        </div>
        {props.subtitle && <div className="__subtitle">{props.subtitle}</div>}
    </div>;

export const DefaultModalList = (props: React.PropsWithChildren<TCommonProps>) =>
    <div className={classnames('__list', props.className)}>
        {props.children}
    </div>;

type TItemProps = TCommonProps & {
    dangerIcon?    : boolean;
    disabled?      : boolean;
    action?        : boolean;
    multipleRows?  : boolean;
    showOnMobile?  : boolean;
    notImplemented?: boolean;
    onClick?       : () => void;
};

export const DefaultModalItem = ({
    dangerIcon     = false,
    disabled       = false,
    action         = false,
    multipleRows   = false,
    showOnMobile   = false,
    notImplemented = false,
    className,
    onClick,
    children
}: React.PropsWithChildren<TItemProps>) =>
    <div className={classnames('__item', className, {
        '--danger-icon': dangerIcon,
        '--disabled': disabled,
        '--action': action,
        '--multiple-rows': multipleRows,
        '--show-on-mobile': showOnMobile,
        '--not-implemented': notImplemented,
    })} onClick={onClick}>
        {children}
    </div>;

type TRowProps = TCommonProps & {
    secondary?: boolean;
    indented?: boolean;
};

export const DefaultModalRow = ({
    indented = false,
    secondary = false,
    className,
    children
}: React.PropsWithChildren<TRowProps>) =>
    <div className={classnames('__row', className, {
        '--indented': indented,
        '--secondary': secondary,
    })}>
        {children}
    </div>;

export const DefaultModalDivider = (props: React.PropsWithChildren<TCommonProps>) =>
    <div className={classnames('flex justify-center my-2', props.className)}>
        <div className="border-t border-gray-500 w-full opacity-40" style={{ height: 1 }}></div>
    </div>