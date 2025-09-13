 import {ac as Wi, ad as Yi, ae as Gi, af as Ki, a1 as cn, ag as Qt, ah as Xi, g as _t, a as Rt, ai as gn, aj as Qi, ak as xn, al as hn, am as Ea, e as Dt, an as Fn, u as Bt, s as xe, b as Se, d as ae, m as it, Y as Ln, a3 as Aa, G as at, H as tt, f as Ba, ao as Zi, ap as Ji, P as Yo, aq as la, ar as Gt, as as io, at as pn, au as Mn, av as Go, aw as Ko, ax as er, ay as tr, Q as $n, az as ar, aA as nr, aB as or, aC as ir, B as K, U as Wt, N as z, aD as an, aE as rr, aF as Na, T as F, aG as lr, aH as ro, aI as sr, aJ as cr, aK as Ia, aL as _a, aM as pr, aN as nn, aO as Xo, aP as dr, aQ as ur, aR as mr, aS as fr, V as Bn, aT as Nn, $ as Qo, aU as vr, aV as gr, a0 as Zo, W as lo, w as Jo, aW as Fe, aa as Ie, aX as nt, aY as xr, aZ as sa, a_ as ca, a$ as ei, b0 as zn, b1 as Xa, c as hr} from "./index-BqrR_uo2.js";
import {j as r, d as br, u as ta, q as Vn} from "./tanstack.Chy1u8Eh.js";
import {r as h, R as g, A as yr, t as wr, M as Cr, f as Sr} from "./agGrid.D1-JvxK1.js";
import {a as Nt, b as zt, c as Hn, d as ti, e as Er, f as ai, A as on, B as bn, g as Ar, h as ni, i as kr, j as oi} from "./api.applications.queries.Ba5jUig4.js";
import {g as Tr} from "./lodash.CkWIuci5.js";
import {p as be, n as De, q as ii, r as Ga, s as za, t as ka, S as Pr} from "./index.esm.Bxu6vCx2.js";
import {u as Ft, b as Ve, c as aa, d as Or, F as Ra, a as Da} from "./formik.esm.Da7QC0oX.js";
import {D as Ot} from "./datetime.C3IbXDYv.js";
import {I as je, b as jr, M as Ir, a as _r} from "./Menu.Cz1w204m.js";
import {n as Qa, H as Rr} from "./handsontable.CVjIQxP-.js";
import {g as ri} from "./getFirstStringFormikError.DpkVVq_l.js";
import {u as Ta} from "./index.8LsZympF.js";
import {P as de} from "./fontawesome.sUQUXoJk.js";
import {P as Dr} from "./PureLink.HuZIq23y.js";
import {u as Fr} from "./useTitle.BqGt3IvK.js";
import "./agCharts.BwH7V64n.js";
try {
    let e = typeof window < "u" ? window : typeof global < "u" ? global : typeof globalThis < "u" ? globalThis : typeof self < "u" ? self : {}
      , t = new e.Error().stack;
    t && (e._sentryDebugIds = e._sentryDebugIds || {},
    e._sentryDebugIds[t] = "4d92626e-551c-4b04-a6f5-75044cfca989",
    e._sentryDebugIdIdentifier = "sentry-dbid-4d92626e-551c-4b04-a6f5-75044cfca989");
} catch {}
function Lr({props: e, name: t}) {
    return Wi({
        props: e,
        name: t,
        defaultTheme: Yi,
        themeId: Gi
    });
}
const Mr = (e) => {
    const t = h.useRef({});
    return h.useEffect( () => {
        t.current = e;
    }
    ),
    t.current;
}
;
function so(e) {
    return e.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
}
function $r(e={}) {
    const {ignoreAccents: t=true, ignoreCase: a=true, limit: n, matchFrom: o="any", stringify: i, trim: l=false} = e;
    return (s, {inputValue: c, getOptionLabel: d}) => {
        let p = l ? c.trim() : c;
        a && (p = p.toLowerCase()),
        t && (p = so(p));
        const u = p ? s.filter( (f) => {
            let m = (i || d)(f);
            return a && (m = m.toLowerCase()),
            t && (m = so(m)),
            o === "start" ? m.startsWith(p) : m.includes(p);
        }
        ) : s;
        return typeof n == "number" ? u.slice(0, n) : u;
    }
    ;
}
const Br = $r()
  , Nr = 5
  , zr = (e) => {
    var _a2;
    return e.current !== null && ((_a2 = e.current.parentElement) == null ? void 0 : _a2.contains(document.activeElement));
}
  , Vr = [];
function co(e, t, a, n) {
    if (t || e == null || n)
        return "";
    const o = a(e);
    return typeof o == "string" ? o : "";
}
function Hr(e) {
    const {unstable_isActiveElementInListbox: t=zr, unstable_classNamePrefix: a="Mui", autoComplete: n=false, autoHighlight: o=false, autoSelect: i=false, blurOnSelect: l=false, clearOnBlur: s=!e.freeSolo, clearOnEscape: c=false, componentName: d="useAutocomplete", defaultValue: p=e.multiple ? Vr : null, disableClearable: u=false, disableCloseOnSelect: f=false, disabled: m, disabledItemsFocusable: b=false, disableListWrap: y=false, filterOptions: w=Br, filterSelectedOptions: E=false, freeSolo: A=false, getOptionDisabled: C, getOptionKey: S, getOptionLabel: T= (x) => x.label ?? x, groupBy: P, handleHomeEndKeys: _=!e.freeSolo, id: U, includeInputInList: L=false, inputValue: $, isOptionEqualToValue: q= (x, v) => x === v, multiple: M=false, onChange: G, onClose: H, onHighlightChange: Z, onInputChange: W, onOpen: le, open: Ee, openOnFocus: ne=false, options: Ae, readOnly: pe=false, renderValue: se, selectOnFocus: Ge=!e.freeSolo, value: ye} = e
      , ce = Ki(U);
    let D = T;
    D = (x) => {
        const v = T(x);
        return typeof v != "string" ? String(v) : v;
    }
    ;
    const he = h.useRef(false)
      , Le = h.useRef(true)
      , ue = h.useRef(null)
      , ke = h.useRef(null)
      , [qe,Qe] = h.useState(null)
      , [Oe,rt] = h.useState(-1)
      , St = o ? 0 : -1
      , me = h.useRef(St)
      , Te = h.useRef(co(p ?? ye, M, D)).current
      , [R,Me] = cn({
        controlled: ye,
        default: p,
        name: d
    })
      , [Y,_e] = cn({
        controlled: $,
        default: Te,
        name: d,
        state: "inputValue"
    })
      , [Ke,lt] = h.useState(false)
      , k = h.useCallback( (x, v, j) => {
        if (!(M ? R.length < v.length : v !== null) && !s)
            return;
        const V = co(v, M, D, se);
        Y !== V && (_e(V),
        W && W(x, V, j));
    }
    , [D, Y, M, W, _e, s, R, se])
      , [I,oe] = cn({
        controlled: Ee,
        default: false,
        name: d,
        state: "open"
    })
      , [fe,ve] = h.useState(true)
      , st = !M && R != null && Y === D(R)
      , we = I && !pe
      , re = we ? w(Ae.filter( (x) => !(E && (M ? R : [R]).some( (v) => v !== null && q(x, v)))), {
        inputValue: st && fe ? "" : Y,
        getOptionLabel: D
    }) : []
      , Be = Mr({
        filteredOptions: re,
        value: R,
        inputValue: Y
    });
    h.useEffect( () => {
        const x = R !== Be.value;
        Ke && !x || A && !x || k(null, R, "reset");
    }
    , [R, k, Ke, Be.value, A]);
    const Et = I && re.length > 0 && !pe
      , $e = Qt( (x) => {
        if (x === -1)
            ue.current.focus();
        else {
            const v = se ? "data-item-index" : "data-tag-index";
            qe.querySelector(`[${v}="${x}"]`).focus();
        }
    }
    );
    h.useEffect( () => {
        M && Oe > R.length - 1 && (rt(-1),
        $e(-1));
    }
    , [R, M, Oe, $e]);
    function ct(x, v) {
        if (!ke.current || x < 0 || x >= re.length)
            return -1;
        let j = x;
        for (; ; ) {
            const Q = ke.current.querySelector(`[data-option-index="${j}"]`)
              , V = b ? false : !Q || Q.disabled || Q.getAttribute("aria-disabled") === "true";
            if (Q && Q.hasAttribute("tabindex") && !V)
                return j;
            if (v === "next" ? j = (j + 1) % re.length : j = (j - 1 + re.length) % re.length,
            j === x)
                return -1;
        }
    }
    const Xe = Qt( ({event: x, index: v, reason: j}) => {
        if (me.current = v,
        v === -1 ? ue.current.removeAttribute("aria-activedescendant") : ue.current.setAttribute("aria-activedescendant", `${ce}-option-${v}`),
        Z && ["mouse", "keyboard", "touch"].includes(j) && Z(x, v === -1 ? null : re[v], j),
        !ke.current)
            return;
        const Q = ke.current.querySelector(`[role="option"].${a}-focused`);
        Q && (Q.classList.remove(`${a}-focused`),
        Q.classList.remove(`${a}-focusVisible`));
        let V = ke.current;
        if (ke.current.getAttribute("role") !== "listbox" && (V = ke.current.parentElement.querySelector('[role="listbox"]')),
        !V)
            return;
        if (v === -1) {
            V.scrollTop = 0;
            return;
        }
        const Pe = ke.current.querySelector(`[data-option-index="${v}"]`);
        if (Pe && (Pe.classList.add(`${a}-focused`),
        j === "keyboard" && Pe.classList.add(`${a}-focusVisible`),
        V.scrollHeight > V.clientHeight && j !== "mouse" && j !== "touch")) {
            const Re = Pe
              , ut = V.clientHeight + V.scrollTop
              , sn = Re.offsetTop + Re.offsetHeight;
            sn > ut ? V.scrollTop = sn - V.clientHeight : Re.offsetTop - Re.offsetHeight * (P ? 1.3 : 0) < V.scrollTop && (V.scrollTop = Re.offsetTop - Re.offsetHeight * (P ? 1.3 : 0));
        }
    }
    )
      , He = Qt( ({event: x, diff: v, direction: j="next", reason: Q}) => {
        if (!we)
            return;
        const Pe = ct(( () => {
            const Re = re.length - 1;
            if (v === "reset")
                return St;
            if (v === "start")
                return 0;
            if (v === "end")
                return Re;
            const ut = me.current + v;
            return ut < 0 ? ut === -1 && L ? -1 : y && me.current !== -1 || Math.abs(v) > 1 ? 0 : Re : ut > Re ? ut === Re + 1 && L ? -1 : y || Math.abs(v) > 1 ? Re : 0 : ut;
        }
        )(), j);
        if (Xe({
            index: Pe,
            reason: Q,
            event: x
        }),
        n && v !== "reset")
            if (Pe === -1)
                ue.current.value = Y;
            else {
                const Re = D(re[Pe]);
                ue.current.value = Re,
                Re.toLowerCase().indexOf(Y.toLowerCase()) === 0 && Y.length > 0 && ue.current.setSelectionRange(Y.length, Re.length);
            }
    }
    )
      , At = () => {
        const x = (v, j) => {
            const Q = v ? D(v) : ""
              , V = j ? D(j) : "";
            return Q === V;
        }
        ;
        if (me.current !== -1 && Be.filteredOptions && Be.filteredOptions.length !== re.length && Be.inputValue === Y && (M ? R.length === Be.value.length && Be.value.every( (v, j) => D(R[j]) === D(v)) : x(Be.value, R))) {
            const v = Be.filteredOptions[me.current];
            if (v)
                return re.findIndex( (j) => D(j) === D(v));
        }
        return -1;
    }
      , pt = h.useCallback( () => {
        if (!we)
            return;
        const x = At();
        if (x !== -1) {
            me.current = x;
            return;
        }
        const v = M ? R[0] : R;
        if (re.length === 0 || v == null) {
            He({
                diff: "reset"
            });
            return;
        }
        if (ke.current) {
            if (v != null) {
                const j = re[me.current];
                if (M && j && R.findIndex( (V) => q(j, V)) !== -1)
                    return;
                const Q = re.findIndex( (V) => q(V, v));
                Q === -1 ? He({
                    diff: "reset"
                }) : Xe({
                    index: Q
                });
                return;
            }
            if (me.current >= re.length - 1) {
                Xe({
                    index: re.length - 1
                });
                return;
            }
            Xe({
                index: me.current
            });
        }
    }
    , [re.length, M ? false : R, E, He, Xe, we, Y, M])
      , ga = Qt( (x) => {
        Xi(ke, x),
        x && pt();
    }
    );
    h.useEffect( () => {
        pt();
    }
    , [pt]);
    const We = (x) => {
        I || (oe(true),
        ve(true),
        le && le(x));
    }
      , yt = (x, v) => {
        I && (oe(false),
        H && H(x, v));
    }
      , wt = (x, v, j, Q) => {
        if (M) {
            if (R.length === v.length && R.every( (V, Pe) => V === v[Pe]))
                return;
        } else if (R === v)
            return;
        G && G(x, v, j, Q),
        Me(v);
    }
      , Vt = h.useRef(false)
      , dt = (x, v, j="selectOption", Q="options") => {
        let V = j
          , Pe = v;
        if (M) {
            Pe = Array.isArray(R) ? R.slice() : [];
            const Re = Pe.findIndex( (ut) => q(v, ut));
            Re === -1 ? Pe.push(v) : Q !== "freeSolo" && (Pe.splice(Re, 1),
            V = "removeOption");
        }
        k(x, Pe, V),
        wt(x, Pe, V, {
            option: v
        }),
        !f && (!x || !x.ctrlKey && !x.metaKey) && yt(x, V),
        (l === true || l === "touch" && Vt.current || l === "mouse" && !Vt.current) && ue.current.blur();
    }
    ;
    function ia(x, v) {
        if (x === -1)
            return -1;
        let j = x;
        for (; ; ) {
            if (v === "next" && j === R.length || v === "previous" && j === -1)
                return -1;
            const Q = se ? "data-item-index" : "data-tag-index"
              , V = qe.querySelector(`[${Q}="${j}"]`);
            if (!V || !V.hasAttribute("tabindex") || V.disabled || V.getAttribute("aria-disabled") === "true")
                j += v === "next" ? 1 : -1;
            else
                return j;
        }
    }
    const ra = (x, v) => {
        if (!M)
            return;
        Y === "" && yt(x, "toggleInput");
        let j = Oe;
        Oe === -1 ? Y === "" && v === "previous" && (j = R.length - 1) : (j += v === "next" ? 1 : -1,
        j < 0 && (j = 0),
        j === R.length && (j = -1)),
        j = ia(j, v),
        rt(j),
        $e(j);
    }
      , Mt = (x) => {
        he.current = true,
        _e(""),
        W && W(x, "", "clear"),
        wt(x, M ? [] : null, "clear");
    }
      , O = (x) => (v) => {
        if (x.onKeyDown && x.onKeyDown(v),
        !v.defaultMuiPrevented && (Oe !== -1 && !["ArrowLeft", "ArrowRight"].includes(v.key) && (rt(-1),
        $e(-1)),
        v.which !== 229))
            switch (v.key) {
            case "Home":
                we && _ && (v.preventDefault(),
                He({
                    diff: "start",
                    direction: "next",
                    reason: "keyboard",
                    event: v
                }));
                break;
            case "End":
                we && _ && (v.preventDefault(),
                He({
                    diff: "end",
                    direction: "previous",
                    reason: "keyboard",
                    event: v
                }));
                break;
            case "PageUp":
                v.preventDefault(),
                He({
                    diff: -5,
                    direction: "previous",
                    reason: "keyboard",
                    event: v
                }),
                We(v);
                break;
            case "PageDown":
                v.preventDefault(),
                He({
                    diff: Nr,
                    direction: "next",
                    reason: "keyboard",
                    event: v
                }),
                We(v);
                break;
            case "ArrowDown":
                v.preventDefault(),
                He({
                    diff: 1,
                    direction: "next",
                    reason: "keyboard",
                    event: v
                }),
                We(v);
                break;
            case "ArrowUp":
                v.preventDefault(),
                He({
                    diff: -1,
                    direction: "previous",
                    reason: "keyboard",
                    event: v
                }),
                We(v);
                break;
            case "ArrowLeft":
                !M && se ? $e(0) : ra(v, "previous");
                break;
            case "ArrowRight":
                !M && se ? $e(-1) : ra(v, "next");
                break;
            case "Enter":
                if (me.current !== -1 && we) {
                    const j = re[me.current]
                      , Q = C ? C(j) : false;
                    if (v.preventDefault(),
                    Q)
                        return;
                    dt(v, j, "selectOption"),
                    n && ue.current.setSelectionRange(ue.current.value.length, ue.current.value.length);
                } else
                    A && Y !== "" && st === false && (M && v.preventDefault(),
                    dt(v, Y, "createOption", "freeSolo"));
                break;
            case "Escape":
                we ? (v.preventDefault(),
                v.stopPropagation(),
                yt(v, "escape")) : c && (Y !== "" || M && R.length > 0 || se) && (v.preventDefault(),
                v.stopPropagation(),
                Mt(v));
                break;
            case "Backspace":
                if (M && !pe && Y === "" && R.length > 0) {
                    const j = Oe === -1 ? R.length - 1 : Oe
                      , Q = R.slice();
                    Q.splice(j, 1),
                    wt(v, Q, "removeOption", {
                        option: R[j]
                    });
                }
                !M && se && !pe && (Me(null),
                $e(-1));
                break;
            case "Delete":
                if (M && !pe && Y === "" && R.length > 0 && Oe !== -1) {
                    const j = Oe
                      , Q = R.slice();
                    Q.splice(j, 1),
                    wt(v, Q, "removeOption", {
                        option: R[j]
                    });
                }
                !M && se && !pe && (Me(null),
                $e(-1));
                break;
            }
    }
      , B = (x) => {
        lt(true),
        ne && !he.current && We(x);
    }
      , J = (x) => {
        if (t(ke)) {
            ue.current.focus();
            return;
        }
        lt(false),
        Le.current = true,
        he.current = false,
        i && me.current !== -1 && we ? dt(x, re[me.current], "blur") : i && A && Y !== "" ? dt(x, Y, "blur", "freeSolo") : s && k(x, R, "blur"),
        yt(x, "blur");
    }
      , ee = (x) => {
        const v = x.target.value;
        Y !== v && (_e(v),
        ve(false),
        W && W(x, v, "input")),
        v === "" ? !u && !M && wt(x, null, "clear") : We(x);
    }
      , ge = (x) => {
        const v = Number(x.currentTarget.getAttribute("data-option-index"));
        me.current !== v && Xe({
            event: x,
            index: v,
            reason: "mouse"
        });
    }
      , Ze = (x) => {
        Xe({
            event: x,
            index: Number(x.currentTarget.getAttribute("data-option-index")),
            reason: "touch"
        }),
        Vt.current = true;
    }
      , Ct = (x) => {
        const v = Number(x.currentTarget.getAttribute("data-option-index"));
        dt(x, re[v], "selectOption"),
        Vt.current = false;
    }
      , kt = (x) => (v) => {
        const j = R.slice();
        j.splice(x, 1),
        wt(v, j, "removeOption", {
            option: R[x]
        });
    }
      , xa = (x) => {
        wt(x, null, "removeOption", {
            option: R
        });
    }
      , Tt = (x) => {
        I ? yt(x, "toggleInput") : We(x);
    }
      , Ne = (x) => {
        x.currentTarget.contains(x.target) && x.target.getAttribute("id") !== ce && x.preventDefault();
    }
      , ze = (x) => {
        x.currentTarget.contains(x.target) && (ue.current.focus(),
        Ge && Le.current && ue.current.selectionEnd - ue.current.selectionStart === 0 && ue.current.select(),
        Le.current = false);
    }
      , Ht = (x) => {
        !m && (Y === "" || !I) && Tt(x);
    }
    ;
    let ha = A && Y.length > 0;
    ha = ha || (M ? R.length > 0 : R !== null);
    let $a = re;
    return P && ($a = re.reduce( (x, v, j) => {
        const Q = P(v);
        return x.length > 0 && x[x.length - 1].group === Q ? x[x.length - 1].options.push(v) : x.push({
            key: j,
            index: j,
            group: Q,
            options: [v]
        }),
        x;
    }
    , [])),
    m && Ke && J(),
    {
        getRootProps: (x={}) => ({
            ...x,
            onKeyDown: O(x),
            onMouseDown: Ne,
            onClick: ze
        }),
        getInputLabelProps: () => ({
            id: `${ce}-label`,
            htmlFor: ce
        }),
        getInputProps: () => ({
            id: ce,
            value: Y,
            onBlur: J,
            onFocus: B,
            onChange: ee,
            onMouseDown: Ht,
            "aria-activedescendant": we ? "" : null,
            "aria-autocomplete": n ? "both" : "list",
            "aria-controls": Et ? `${ce}-listbox` : void 0,
            "aria-expanded": Et,
            autoComplete: "off",
            ref: ue,
            autoCapitalize: "none",
            spellCheck: "false",
            role: "combobox",
            disabled: m
        }),
        getClearProps: () => ({
            tabIndex: -1,
            type: "button",
            onClick: Mt
        }),
        getItemProps: ({index: x=0}={}) => ({
            ...M && {
                key: x
            },
            ...se ? {
                "data-item-index": x
            } : {
                "data-tag-index": x
            },
            tabIndex: -1,
            ...!pe && {
                onDelete: M ? kt(x) : xa
            }
        }),
        getPopupIndicatorProps: () => ({
            tabIndex: -1,
            type: "button",
            onClick: Tt
        }),
        getTagProps: ({index: x}) => ({
            key: x,
            "data-tag-index": x,
            tabIndex: -1,
            ...!pe && {
                onDelete: kt(x)
            }
        }),
        getListboxProps: () => ({
            role: "listbox",
            id: `${ce}-listbox`,
            "aria-labelledby": `${ce}-label`,
            ref: ga,
            onMouseDown: (x) => {
                x.preventDefault();
            }
        }),
        getOptionProps: ({index: x, option: v}) => {
            const j = (M ? R : [R]).some( (V) => V != null && q(v, V))
              , Q = C ? C(v) : false;
            return {
                key: (S == null ? void 0 : S(v)) ?? D(v),
                tabIndex: -1,
                role: "option",
                id: `${ce}-option-${x}`,
                onMouseMove: ge,
                onClick: Ct,
                onTouchStart: Ze,
                "data-option-index": x,
                "aria-disabled": Q,
                "aria-selected": j
            };
        }
        ,
        id: ce,
        inputValue: Y,
        value: R,
        dirty: ha,
        expanded: we && qe,
        popupOpen: we,
        focused: Ke || Oe !== -1,
        anchorEl: qe,
        setAnchorEl: Qe,
        focusedItem: Oe,
        focusedTag: Oe,
        groupedOptions: $a
    };
}
var Je = "top"
  , ht = "bottom"
  , bt = "right"
  , et = "left"
  , Un = "auto"
  , Fa = [Je, ht, bt, et]
  , pa = "start"
  , Pa = "end"
  , Ur = "clippingParents"
  , li = "viewport"
  , ya = "popper"
  , qr = "reference"
  , po = Fa.reduce(function(e, t) {
    return e.concat([t + "-" + pa, t + "-" + Pa]);
}, [])
  , si = [].concat(Fa, [Un]).reduce(function(e, t) {
    return e.concat([t, t + "-" + pa, t + "-" + Pa]);
}, [])
  , Wr = "beforeRead"
  , Yr = "read"
  , Gr = "afterRead"
  , Kr = "beforeMain"
  , Xr = "main"
  , Qr = "afterMain"
  , Zr = "beforeWrite"
  , Jr = "write"
  , el = "afterWrite"
  , tl = [Wr, Yr, Gr, Kr, Xr, Qr, Zr, Jr, el];
function It(e) {
    return e ? (e.nodeName || "").toLowerCase() : null;
}
function ot(e) {
    if (e == null)
        return window;
    if (e.toString() !== "[object Window]") {
        var t = e.ownerDocument;
        return t && t.defaultView || window;
    }
    return e;
}
function Jt(e) {
    var t = ot(e).Element;
    return e instanceof t || e instanceof Element;
}
function vt(e) {
    var t = ot(e).HTMLElement;
    return e instanceof t || e instanceof HTMLElement;
}
function qn(e) {
    if (typeof ShadowRoot > "u")
        return false;
    var t = ot(e).ShadowRoot;
    return e instanceof t || e instanceof ShadowRoot;
}
function al(e) {
    var t = e.state;
    Object.keys(t.elements).forEach(function(a) {
        var n = t.styles[a] || {}
          , o = t.attributes[a] || {}
          , i = t.elements[a];
        !vt(i) || !It(i) || (Object.assign(i.style, n),
        Object.keys(o).forEach(function(l) {
            var s = o[l];
            s === false ? i.removeAttribute(l) : i.setAttribute(l, s === true ? "" : s);
        }));
    });
}
function nl(e) {
    var t = e.state
      , a = {
        popper: {
            position: t.options.strategy,
            left: "0",
            top: "0",
            margin: "0"
        },
        arrow: {
            position: "absolute"
        },
        reference: {}
    };
    return Object.assign(t.elements.popper.style, a.popper),
    t.styles = a,
    t.elements.arrow && Object.assign(t.elements.arrow.style, a.arrow),
    function() {
        Object.keys(t.elements).forEach(function(n) {
            var o = t.elements[n]
              , i = t.attributes[n] || {}
              , l = Object.keys(t.styles.hasOwnProperty(n) ? t.styles[n] : a[n])
              , s = l.reduce(function(c, d) {
                return c[d] = "",
                c;
            }, {});
            !vt(o) || !It(o) || (Object.assign(o.style, s),
            Object.keys(i).forEach(function(c) {
                o.removeAttribute(c);
            }));
        });
    }
    ;
}
const ol = {
    name: "applyStyles",
    enabled: true,
    phase: "write",
    fn: al,
    effect: nl,
    requires: ["computeStyles"]
};
function jt(e) {
    return e.split("-")[0];
}
var Zt = Math.max
  , Za = Math.min
  , da = Math.round;
function yn() {
    var e = navigator.userAgentData;
    return e != null && e.brands && Array.isArray(e.brands) ? e.brands.map(function(t) {
        return t.brand + "/" + t.version;
    }).join(" ") : navigator.userAgent;
}
function ci() {
    return !/^((?!chrome|android).)*safari/i.test(yn());
}
function ua(e, t, a) {
    t === void 0 && (t = false),
    a === void 0 && (a = false);
    var n = e.getBoundingClientRect()
      , o = 1
      , i = 1;
    t && vt(e) && (o = e.offsetWidth > 0 && da(n.width) / e.offsetWidth || 1,
    i = e.offsetHeight > 0 && da(n.height) / e.offsetHeight || 1);
    var l = Jt(e) ? ot(e) : window
      , s = l.visualViewport
      , c = !ci() && a
      , d = (n.left + (c && s ? s.offsetLeft : 0)) / o
      , p = (n.top + (c && s ? s.offsetTop : 0)) / i
      , u = n.width / o
      , f = n.height / i;
    return {
        width: u,
        height: f,
        top: p,
        right: d + u,
        bottom: p + f,
        left: d,
        x: d,
        y: p
    };
}
function Wn(e) {
    var t = ua(e)
      , a = e.offsetWidth
      , n = e.offsetHeight;
    return Math.abs(t.width - a) <= 1 && (a = t.width),
    Math.abs(t.height - n) <= 1 && (n = t.height),
    {
        x: e.offsetLeft,
        y: e.offsetTop,
        width: a,
        height: n
    };
}
function pi(e, t) {
    var a = t.getRootNode && t.getRootNode();
    if (e.contains(t))
        return true;
    if (a && qn(a)) {
        var n = t;
        do {
            if (n && e.isSameNode(n))
                return true;
            n = n.parentNode || n.host;
        } while (n);
    }
    return false;
}
function $t(e) {
    return ot(e).getComputedStyle(e);
}
function il(e) {
    return ["table", "td", "th"].indexOf(It(e)) >= 0;
}
function Yt(e) {
    return ((Jt(e) ? e.ownerDocument : e.document) || window.document).documentElement;
}
function rn(e) {
    return It(e) === "html" ? e : e.assignedSlot || e.parentNode || (qn(e) ? e.host : null) || Yt(e);
}
function uo(e) {
    return !vt(e) || $t(e).position === "fixed" ? null : e.offsetParent;
}
function rl(e) {
    var t = /firefox/i.test(yn())
      , a = /Trident/i.test(yn());
    if (a && vt(e)) {
        var n = $t(e);
        if (n.position === "fixed")
            return null;
    }
    var o = rn(e);
    for (qn(o) && (o = o.host); vt(o) && ["html", "body"].indexOf(It(o)) < 0; ) {
        var i = $t(o);
        if (i.transform !== "none" || i.perspective !== "none" || i.contain === "paint" || ["transform", "perspective"].indexOf(i.willChange) !== -1 || t && i.willChange === "filter" || t && i.filter && i.filter !== "none")
            return o;
        o = o.parentNode;
    }
    return null;
}
function La(e) {
    for (var t = ot(e), a = uo(e); a && il(a) && $t(a).position === "static"; )
        a = uo(a);
    return a && (It(a) === "html" || It(a) === "body" && $t(a).position === "static") ? t : a || rl(e) || t;
}
function Yn(e) {
    return ["top", "bottom"].indexOf(e) >= 0 ? "x" : "y";
}
function Ca(e, t, a) {
    return Zt(e, Za(t, a));
}
function ll(e, t, a) {
    var n = Ca(e, t, a);
    return n > a ? a : n;
}
function di() {
    return {
        top: 0,
        right: 0,
        bottom: 0,
        left: 0
    };
}
function ui(e) {
    return Object.assign({}, di(), e);
}
function mi(e, t) {
    return t.reduce(function(a, n) {
        return a[n] = e,
        a;
    }, {});
}
var sl = function(t, a) {
    return t = typeof t == "function" ? t(Object.assign({}, a.rects, {
        placement: a.placement
    })) : t,
    ui(typeof t != "number" ? t : mi(t, Fa));
};
function cl(e) {
    var t, a = e.state, n = e.name, o = e.options, i = a.elements.arrow, l = a.modifiersData.popperOffsets, s = jt(a.placement), c = Yn(s), d = [et, bt].indexOf(s) >= 0, p = d ? "height" : "width";
    if (!(!i || !l)) {
        var u = sl(o.padding, a)
          , f = Wn(i)
          , m = c === "y" ? Je : et
          , b = c === "y" ? ht : bt
          , y = a.rects.reference[p] + a.rects.reference[c] - l[c] - a.rects.popper[p]
          , w = l[c] - a.rects.reference[c]
          , E = La(i)
          , A = E ? c === "y" ? E.clientHeight || 0 : E.clientWidth || 0 : 0
          , C = y / 2 - w / 2
          , S = u[m]
          , T = A - f[p] - u[b]
          , P = A / 2 - f[p] / 2 + C
          , _ = Ca(S, P, T)
          , U = c;
        a.modifiersData[n] = (t = {},
        t[U] = _,
        t.centerOffset = _ - P,
        t);
    }
}
function pl(e) {
    var t = e.state
      , a = e.options
      , n = a.element
      , o = n === void 0 ? "[data-popper-arrow]" : n;
    o != null && (typeof o == "string" && (o = t.elements.popper.querySelector(o),
    !o) || pi(t.elements.popper, o) && (t.elements.arrow = o));
}
const dl = {
    name: "arrow",
    enabled: true,
    phase: "main",
    fn: cl,
    effect: pl,
    requires: ["popperOffsets"],
    requiresIfExists: ["preventOverflow"]
};
function ma(e) {
    return e.split("-")[1];
}
var ul = {
    top: "auto",
    right: "auto",
    bottom: "auto",
    left: "auto"
};
function ml(e, t) {
    var a = e.x
      , n = e.y
      , o = t.devicePixelRatio || 1;
    return {
        x: da(a * o) / o || 0,
        y: da(n * o) / o || 0
    };
}
function mo(e) {
    var t, a = e.popper, n = e.popperRect, o = e.placement, i = e.variation, l = e.offsets, s = e.position, c = e.gpuAcceleration, d = e.adaptive, p = e.roundOffsets, u = e.isFixed, f = l.x, m = f === void 0 ? 0 : f, b = l.y, y = b === void 0 ? 0 : b, w = typeof p == "function" ? p({
        x: m,
        y
    }) : {
        x: m,
        y
    };
    m = w.x,
    y = w.y;
    var E = l.hasOwnProperty("x")
      , A = l.hasOwnProperty("y")
      , C = et
      , S = Je
      , T = window;
    if (d) {
        var P = La(a)
          , _ = "clientHeight"
          , U = "clientWidth";
        if (P === ot(a) && (P = Yt(a),
        $t(P).position !== "static" && s === "absolute" && (_ = "scrollHeight",
        U = "scrollWidth")),
        P = P,
        o === Je || (o === et || o === bt) && i === Pa) {
            S = ht;
            var L = u && P === T && T.visualViewport ? T.visualViewport.height : P[_];
            y -= L - n.height,
            y *= c ? 1 : -1;
        }
        if (o === et || (o === Je || o === ht) && i === Pa) {
            C = bt;
            var $ = u && P === T && T.visualViewport ? T.visualViewport.width : P[U];
            m -= $ - n.width,
            m *= c ? 1 : -1;
        }
    }
    var q = Object.assign({
        position: s
    }, d && ul)
      , M = p === true ? ml({
        x: m,
        y
    }, ot(a)) : {
        x: m,
        y
    };
    if (m = M.x,
    y = M.y,
    c) {
        var G;
        return Object.assign({}, q, (G = {},
        G[S] = A ? "0" : "",
        G[C] = E ? "0" : "",
        G.transform = (T.devicePixelRatio || 1) <= 1 ? "translate(" + m + "px, " + y + "px)" : "translate3d(" + m + "px, " + y + "px, 0)",
        G));
    }
    return Object.assign({}, q, (t = {},
    t[S] = A ? y + "px" : "",
    t[C] = E ? m + "px" : "",
    t.transform = "",
    t));
}
function fl(e) {
    var t = e.state
      , a = e.options
      , n = a.gpuAcceleration
      , o = n === void 0 ? true : n
      , i = a.adaptive
      , l = i === void 0 ? true : i
      , s = a.roundOffsets
      , c = s === void 0 ? true : s
      , d = {
        placement: jt(t.placement),
        variation: ma(t.placement),
        popper: t.elements.popper,
        popperRect: t.rects.popper,
        gpuAcceleration: o,
        isFixed: t.options.strategy === "fixed"
    };
    t.modifiersData.popperOffsets != null && (t.styles.popper = Object.assign({}, t.styles.popper, mo(Object.assign({}, d, {
        offsets: t.modifiersData.popperOffsets,
        position: t.options.strategy,
        adaptive: l,
        roundOffsets: c
    })))),
    t.modifiersData.arrow != null && (t.styles.arrow = Object.assign({}, t.styles.arrow, mo(Object.assign({}, d, {
        offsets: t.modifiersData.arrow,
        position: "absolute",
        adaptive: false,
        roundOffsets: c
    })))),
    t.attributes.popper = Object.assign({}, t.attributes.popper, {
        "data-popper-placement": t.placement
    });
}
const vl = {
    name: "computeStyles",
    enabled: true,
    phase: "beforeWrite",
    fn: fl,
    data: {}
};
var Va = {
    passive: true
};
function gl(e) {
    var t = e.state
      , a = e.instance
      , n = e.options
      , o = n.scroll
      , i = o === void 0 ? true : o
      , l = n.resize
      , s = l === void 0 ? true : l
      , c = ot(t.elements.popper)
      , d = [].concat(t.scrollParents.reference, t.scrollParents.popper);
    return i && d.forEach(function(p) {
        p.addEventListener("scroll", a.update, Va);
    }),
    s && c.addEventListener("resize", a.update, Va),
    function() {
        i && d.forEach(function(p) {
            p.removeEventListener("scroll", a.update, Va);
        }),
        s && c.removeEventListener("resize", a.update, Va);
    }
    ;
}
const xl = {
    name: "eventListeners",
    enabled: true,
    phase: "write",
    fn: function() {},
    effect: gl,
    data: {}
};
var hl = {
    left: "right",
    right: "left",
    bottom: "top",
    top: "bottom"
};
function Ka(e) {
    return e.replace(/left|right|bottom|top/g, function(t) {
        return hl[t];
    });
}
var bl = {
    start: "end",
    end: "start"
};
function fo(e) {
    return e.replace(/start|end/g, function(t) {
        return bl[t];
    });
}
function Gn(e) {
    var t = ot(e)
      , a = t.pageXOffset
      , n = t.pageYOffset;
    return {
        scrollLeft: a,
        scrollTop: n
    };
}
function Kn(e) {
    return ua(Yt(e)).left + Gn(e).scrollLeft;
}
function yl(e, t) {
    var a = ot(e)
      , n = Yt(e)
      , o = a.visualViewport
      , i = n.clientWidth
      , l = n.clientHeight
      , s = 0
      , c = 0;
    if (o) {
        i = o.width,
        l = o.height;
        var d = ci();
        (d || !d && t === "fixed") && (s = o.offsetLeft,
        c = o.offsetTop);
    }
    return {
        width: i,
        height: l,
        x: s + Kn(e),
        y: c
    };
}
function wl(e) {
    var t, a = Yt(e), n = Gn(e), o = (t = e.ownerDocument) == null ? void 0 : t.body, i = Zt(a.scrollWidth, a.clientWidth, o ? o.scrollWidth : 0, o ? o.clientWidth : 0), l = Zt(a.scrollHeight, a.clientHeight, o ? o.scrollHeight : 0, o ? o.clientHeight : 0), s = -n.scrollLeft + Kn(e), c = -n.scrollTop;
    return $t(o || a).direction === "rtl" && (s += Zt(a.clientWidth, o ? o.clientWidth : 0) - i),
    {
        width: i,
        height: l,
        x: s,
        y: c
    };
}
function Xn(e) {
    var t = $t(e)
      , a = t.overflow
      , n = t.overflowX
      , o = t.overflowY;
    return /auto|scroll|overlay|hidden/.test(a + o + n);
}
function fi(e) {
    return ["html", "body", "#document"].indexOf(It(e)) >= 0 ? e.ownerDocument.body : vt(e) && Xn(e) ? e : fi(rn(e));
}
function Sa(e, t) {
    var a;
    t === void 0 && (t = []);
    var n = fi(e)
      , o = n === ((a = e.ownerDocument) == null ? void 0 : a.body)
      , i = ot(n)
      , l = o ? [i].concat(i.visualViewport || [], Xn(n) ? n : []) : n
      , s = t.concat(l);
    return o ? s : s.concat(Sa(rn(l)));
}
function wn(e) {
    return Object.assign({}, e, {
        left: e.x,
        top: e.y,
        right: e.x + e.width,
        bottom: e.y + e.height
    });
}
function Cl(e, t) {
    var a = ua(e, false, t === "fixed");
    return a.top = a.top + e.clientTop,
    a.left = a.left + e.clientLeft,
    a.bottom = a.top + e.clientHeight,
    a.right = a.left + e.clientWidth,
    a.width = e.clientWidth,
    a.height = e.clientHeight,
    a.x = a.left,
    a.y = a.top,
    a;
}
function vo(e, t, a) {
    return t === li ? wn(yl(e, a)) : Jt(t) ? Cl(t, a) : wn(wl(Yt(e)));
}
function Sl(e) {
    var t = Sa(rn(e))
      , a = ["absolute", "fixed"].indexOf($t(e).position) >= 0
      , n = a && vt(e) ? La(e) : e;
    return Jt(n) ? t.filter(function(o) {
        return Jt(o) && pi(o, n) && It(o) !== "body";
    }) : [];
}
function El(e, t, a, n) {
    var o = t === "clippingParents" ? Sl(e) : [].concat(t)
      , i = [].concat(o, [a])
      , l = i[0]
      , s = i.reduce(function(c, d) {
        var p = vo(e, d, n);
        return c.top = Zt(p.top, c.top),
        c.right = Za(p.right, c.right),
        c.bottom = Za(p.bottom, c.bottom),
        c.left = Zt(p.left, c.left),
        c;
    }, vo(e, l, n));
    return s.width = s.right - s.left,
    s.height = s.bottom - s.top,
    s.x = s.left,
    s.y = s.top,
    s;
}
function vi(e) {
    var t = e.reference, a = e.element, n = e.placement, o = n ? jt(n) : null, i = n ? ma(n) : null, l = t.x + t.width / 2 - a.width / 2, s = t.y + t.height / 2 - a.height / 2, c;
    switch (o) {
    case Je:
        c = {
            x: l,
            y: t.y - a.height
        };
        break;
    case ht:
        c = {
            x: l,
            y: t.y + t.height
        };
        break;
    case bt:
        c = {
            x: t.x + t.width,
            y: s
        };
        break;
    case et:
        c = {
            x: t.x - a.width,
            y: s
        };
        break;
    default:
        c = {
            x: t.x,
            y: t.y
        };
    }
    var d = o ? Yn(o) : null;
    if (d != null) {
        var p = d === "y" ? "height" : "width";
        switch (i) {
        case pa:
            c[d] = c[d] - (t[p] / 2 - a[p] / 2);
            break;
        case Pa:
            c[d] = c[d] + (t[p] / 2 - a[p] / 2);
            break;
        }
    }
    return c;
}
function Oa(e, t) {
    t === void 0 && (t = {});
    var a = t
      , n = a.placement
      , o = n === void 0 ? e.placement : n
      , i = a.strategy
      , l = i === void 0 ? e.strategy : i
      , s = a.boundary
      , c = s === void 0 ? Ur : s
      , d = a.rootBoundary
      , p = d === void 0 ? li : d
      , u = a.elementContext
      , f = u === void 0 ? ya : u
      , m = a.altBoundary
      , b = m === void 0 ? false : m
      , y = a.padding
      , w = y === void 0 ? 0 : y
      , E = ui(typeof w != "number" ? w : mi(w, Fa))
      , A = f === ya ? qr : ya
      , C = e.rects.popper
      , S = e.elements[b ? A : f]
      , T = El(Jt(S) ? S : S.contextElement || Yt(e.elements.popper), c, p, l)
      , P = ua(e.elements.reference)
      , _ = vi({
        reference: P,
        element: C,
        placement: o
    })
      , U = wn(Object.assign({}, C, _))
      , L = f === ya ? U : P
      , $ = {
        top: T.top - L.top + E.top,
        bottom: L.bottom - T.bottom + E.bottom,
        left: T.left - L.left + E.left,
        right: L.right - T.right + E.right
    }
      , q = e.modifiersData.offset;
    if (f === ya && q) {
        var M = q[o];
        Object.keys($).forEach(function(G) {
            var H = [bt, ht].indexOf(G) >= 0 ? 1 : -1
              , Z = [Je, ht].indexOf(G) >= 0 ? "y" : "x";
            $[G] += M[Z] * H;
        });
    }
    return $;
}
function Al(e, t) {
    t === void 0 && (t = {});
    var a = t
      , n = a.placement
      , o = a.boundary
      , i = a.rootBoundary
      , l = a.padding
      , s = a.flipVariations
      , c = a.allowedAutoPlacements
      , d = c === void 0 ? si : c
      , p = ma(n)
      , u = p ? s ? po : po.filter(function(b) {
        return ma(b) === p;
    }) : Fa
      , f = u.filter(function(b) {
        return d.indexOf(b) >= 0;
    });
    f.length === 0 && (f = u);
    var m = f.reduce(function(b, y) {
        return b[y] = Oa(e, {
            placement: y,
            boundary: o,
            rootBoundary: i,
            padding: l
        })[jt(y)],
        b;
    }, {});
    return Object.keys(m).sort(function(b, y) {
        return m[b] - m[y];
    });
}
function kl(e) {
    if (jt(e) === Un)
        return [];
    var t = Ka(e);
    return [fo(e), t, fo(t)];
}
function Tl(e) {
    var t = e.state
      , a = e.options
      , n = e.name;
    if (!t.modifiersData[n]._skip) {
        for (var o = a.mainAxis, i = o === void 0 ? true : o, l = a.altAxis, s = l === void 0 ? true : l, c = a.fallbackPlacements, d = a.padding, p = a.boundary, u = a.rootBoundary, f = a.altBoundary, m = a.flipVariations, b = m === void 0 ? true : m, y = a.allowedAutoPlacements, w = t.options.placement, E = jt(w), A = E === w, C = c || (A || !b ? [Ka(w)] : kl(w)), S = [w].concat(C).reduce(function(ye, ce) {
            return ye.concat(jt(ce) === Un ? Al(t, {
                placement: ce,
                boundary: p,
                rootBoundary: u,
                padding: d,
                flipVariations: b,
                allowedAutoPlacements: y
            }) : ce);
        }, []), T = t.rects.reference, P = t.rects.popper, _ = /* @__PURE__ */
        new Map(), U = true, L = S[0], $ = 0; $ < S.length; $++) {
            var q = S[$]
              , M = jt(q)
              , G = ma(q) === pa
              , H = [Je, ht].indexOf(M) >= 0
              , Z = H ? "width" : "height"
              , W = Oa(t, {
                placement: q,
                boundary: p,
                rootBoundary: u,
                altBoundary: f,
                padding: d
            })
              , le = H ? G ? bt : et : G ? ht : Je;
            T[Z] > P[Z] && (le = Ka(le));
            var Ee = Ka(le)
              , ne = [];
            if (i && ne.push(W[M] <= 0),
            s && ne.push(W[le] <= 0, W[Ee] <= 0),
            ne.every(function(ye) {
                return ye;
            })) {
                L = q,
                U = false;
                break;
            }
            _.set(q, ne);
        }
        if (U)
            for (var Ae = b ? 3 : 1, pe = function(ce) {
                var D = S.find(function(he) {
                    var Le = _.get(he);
                    if (Le)
                        return Le.slice(0, ce).every(function(ue) {
                            return ue;
                        });
                });
                if (D)
                    return L = D,
                    "break";
            }, se = Ae; se > 0; se--) {
                var Ge = pe(se);
                if (Ge === "break")
                    break;
            }
        t.placement !== L && (t.modifiersData[n]._skip = true,
        t.placement = L,
        t.reset = true);
    }
}
const Pl = {
    name: "flip",
    enabled: true,
    phase: "main",
    fn: Tl,
    requiresIfExists: ["offset"],
    data: {
        _skip: false
    }
};
function go(e, t, a) {
    return a === void 0 && (a = {
        x: 0,
        y: 0
    }),
    {
        top: e.top - t.height - a.y,
        right: e.right - t.width + a.x,
        bottom: e.bottom - t.height + a.y,
        left: e.left - t.width - a.x
    };
}
function xo(e) {
    return [Je, bt, ht, et].some(function(t) {
        return e[t] >= 0;
    });
}
function Ol(e) {
    var t = e.state
      , a = e.name
      , n = t.rects.reference
      , o = t.rects.popper
      , i = t.modifiersData.preventOverflow
      , l = Oa(t, {
        elementContext: "reference"
    })
      , s = Oa(t, {
        altBoundary: true
    })
      , c = go(l, n)
      , d = go(s, o, i)
      , p = xo(c)
      , u = xo(d);
    t.modifiersData[a] = {
        referenceClippingOffsets: c,
        popperEscapeOffsets: d,
        isReferenceHidden: p,
        hasPopperEscaped: u
    },
    t.attributes.popper = Object.assign({}, t.attributes.popper, {
        "data-popper-reference-hidden": p,
        "data-popper-escaped": u
    });
}
const jl = {
    name: "hide",
    enabled: true,
    phase: "main",
    requiresIfExists: ["preventOverflow"],
    fn: Ol
};
function Il(e, t, a) {
    var n = jt(e)
      , o = [et, Je].indexOf(n) >= 0 ? -1 : 1
      , i = typeof a == "function" ? a(Object.assign({}, t, {
        placement: e
    })) : a
      , l = i[0]
      , s = i[1];
    return l = l || 0,
    s = (s || 0) * o,
    [et, bt].indexOf(n) >= 0 ? {
        x: s,
        y: l
    } : {
        x: l,
        y: s
    };
}
function _l(e) {
    var t = e.state
      , a = e.options
      , n = e.name
      , o = a.offset
      , i = o === void 0 ? [0, 0] : o
      , l = si.reduce(function(p, u) {
        return p[u] = Il(u, t.rects, i),
        p;
    }, {})
      , s = l[t.placement]
      , c = s.x
      , d = s.y;
    t.modifiersData.popperOffsets != null && (t.modifiersData.popperOffsets.x += c,
    t.modifiersData.popperOffsets.y += d),
    t.modifiersData[n] = l;
}
const Rl = {
    name: "offset",
    enabled: true,
    phase: "main",
    requires: ["popperOffsets"],
    fn: _l
};
function Dl(e) {
    var t = e.state
      , a = e.name;
    t.modifiersData[a] = vi({
        reference: t.rects.reference,
        element: t.rects.popper,
        placement: t.placement
    });
}
const Fl = {
    name: "popperOffsets",
    enabled: true,
    phase: "read",
    fn: Dl,
    data: {}
};
function Ll(e) {
    return e === "x" ? "y" : "x";
}
function Ml(e) {
    var t = e.state
      , a = e.options
      , n = e.name
      , o = a.mainAxis
      , i = o === void 0 ? true : o
      , l = a.altAxis
      , s = l === void 0 ? false : l
      , c = a.boundary
      , d = a.rootBoundary
      , p = a.altBoundary
      , u = a.padding
      , f = a.tether
      , m = f === void 0 ? true : f
      , b = a.tetherOffset
      , y = b === void 0 ? 0 : b
      , w = Oa(t, {
        boundary: c,
        rootBoundary: d,
        padding: u,
        altBoundary: p
    })
      , E = jt(t.placement)
      , A = ma(t.placement)
      , C = !A
      , S = Yn(E)
      , T = Ll(S)
      , P = t.modifiersData.popperOffsets
      , _ = t.rects.reference
      , U = t.rects.popper
      , L = typeof y == "function" ? y(Object.assign({}, t.rects, {
        placement: t.placement
    })) : y
      , $ = typeof L == "number" ? {
        mainAxis: L,
        altAxis: L
    } : Object.assign({
        mainAxis: 0,
        altAxis: 0
    }, L)
      , q = t.modifiersData.offset ? t.modifiersData.offset[t.placement] : null
      , M = {
        x: 0,
        y: 0
    };
    if (P) {
        if (i) {
            var G, H = S === "y" ? Je : et, Z = S === "y" ? ht : bt, W = S === "y" ? "height" : "width", le = P[S], Ee = le + w[H], ne = le - w[Z], Ae = m ? -U[W] / 2 : 0, pe = A === pa ? _[W] : U[W], se = A === pa ? -U[W] : -_[W], Ge = t.elements.arrow, ye = m && Ge ? Wn(Ge) : {
                width: 0,
                height: 0
            }, ce = t.modifiersData["arrow#persistent"] ? t.modifiersData["arrow#persistent"].padding : di(), D = ce[H], he = ce[Z], Le = Ca(0, _[W], ye[W]), ue = C ? _[W] / 2 - Ae - Le - D - $.mainAxis : pe - Le - D - $.mainAxis, ke = C ? -_[W] / 2 + Ae + Le + he + $.mainAxis : se + Le + he + $.mainAxis, qe = t.elements.arrow && La(t.elements.arrow), Qe = qe ? S === "y" ? qe.clientTop || 0 : qe.clientLeft || 0 : 0, Oe = (G = q == null ? void 0 : q[S]) != null ? G : 0, rt = le + ue - Oe - Qe, St = le + ke - Oe, me = Ca(m ? Za(Ee, rt) : Ee, le, m ? Zt(ne, St) : ne);
            P[S] = me,
            M[S] = me - le;
        }
        if (s) {
            var Te, R = S === "x" ? Je : et, Me = S === "x" ? ht : bt, Y = P[T], _e = T === "y" ? "height" : "width", Ke = Y + w[R], lt = Y - w[Me], k = [Je, et].indexOf(E) !== -1, I = (Te = q == null ? void 0 : q[T]) != null ? Te : 0, oe = k ? Ke : Y - _[_e] - U[_e] - I + $.altAxis, fe = k ? Y + _[_e] + U[_e] - I - $.altAxis : lt, ve = m && k ? ll(oe, Y, fe) : Ca(m ? oe : Ke, Y, m ? fe : lt);
            P[T] = ve,
            M[T] = ve - Y;
        }
        t.modifiersData[n] = M;
    }
}
const $l = {
    name: "preventOverflow",
    enabled: true,
    phase: "main",
    fn: Ml,
    requiresIfExists: ["offset"]
};
function Bl(e) {
    return {
        scrollLeft: e.scrollLeft,
        scrollTop: e.scrollTop
    };
}
function Nl(e) {
    return e === ot(e) || !vt(e) ? Gn(e) : Bl(e);
}
function zl(e) {
    var t = e.getBoundingClientRect()
      , a = da(t.width) / e.offsetWidth || 1
      , n = da(t.height) / e.offsetHeight || 1;
    return a !== 1 || n !== 1;
}
function Vl(e, t, a) {
    a === void 0 && (a = false);
    var n = vt(t)
      , o = vt(t) && zl(t)
      , i = Yt(t)
      , l = ua(e, o, a)
      , s = {
        scrollLeft: 0,
        scrollTop: 0
    }
      , c = {
        x: 0,
        y: 0
    };
    return (n || !n && !a) && ((It(t) !== "body" || Xn(i)) && (s = Nl(t)),
    vt(t) ? (c = ua(t, true),
    c.x += t.clientLeft,
    c.y += t.clientTop) : i && (c.x = Kn(i))),
    {
        x: l.left + s.scrollLeft - c.x,
        y: l.top + s.scrollTop - c.y,
        width: l.width,
        height: l.height
    };
}
function Hl(e) {
    var t = /* @__PURE__ */
    new Map()
      , a = /* @__PURE__ */
    new Set()
      , n = [];
    e.forEach(function(i) {
        t.set(i.name, i);
    });
    function o(i) {
        a.add(i.name);
        var l = [].concat(i.requires || [], i.requiresIfExists || []);
        l.forEach(function(s) {
            if (!a.has(s)) {
                var c = t.get(s);
                c && o(c);
            }
        }),
        n.push(i);
    }
    return e.forEach(function(i) {
        a.has(i.name) || o(i);
    }),
    n;
}
function Ul(e) {
    var t = Hl(e);
    return tl.reduce(function(a, n) {
        return a.concat(t.filter(function(o) {
            return o.phase === n;
        }));
    }, []);
}
function ql(e) {
    var t;
    return function() {
        return t || (t = new Promise(function(a) {
            Promise.resolve().then(function() {
                t = void 0,
                a(e());
            });
        }
        )),
        t;
    }
    ;
}
function Wl(e) {
    var t = e.reduce(function(a, n) {
        var o = a[n.name];
        return a[n.name] = o ? Object.assign({}, o, n, {
            options: Object.assign({}, o.options, n.options),
            data: Object.assign({}, o.data, n.data)
        }) : n,
        a;
    }, {});
    return Object.keys(t).map(function(a) {
        return t[a];
    });
}
var ho = {
    placement: "bottom",
    modifiers: [],
    strategy: "absolute"
};
function bo() {
    for (var e = arguments.length, t = new Array(e), a = 0; a < e; a++)
        t[a] = arguments[a];
    return !t.some(function(n) {
        return !(n && typeof n.getBoundingClientRect == "function");
    });
}
function Yl(e) {
    e === void 0 && (e = {});
    var t = e
      , a = t.defaultModifiers
      , n = a === void 0 ? [] : a
      , o = t.defaultOptions
      , i = o === void 0 ? ho : o;
    return function(s, c, d) {
        d === void 0 && (d = i);
        var p = {
            placement: "bottom",
            orderedModifiers: [],
            options: Object.assign({}, ho, i),
            modifiersData: {},
            elements: {
                reference: s,
                popper: c
            },
            attributes: {},
            styles: {}
        }
          , u = []
          , f = false
          , m = {
            state: p,
            setOptions: function(E) {
                var A = typeof E == "function" ? E(p.options) : E;
                y(),
                p.options = Object.assign({}, i, p.options, A),
                p.scrollParents = {
                    reference: Jt(s) ? Sa(s) : s.contextElement ? Sa(s.contextElement) : [],
                    popper: Sa(c)
                };
                var C = Ul(Wl([].concat(n, p.options.modifiers)));
                return p.orderedModifiers = C.filter(function(S) {
                    return S.enabled;
                }),
                b(),
                m.update();
            },
            forceUpdate: function() {
                if (!f) {
                    var E = p.elements
                      , A = E.reference
                      , C = E.popper;
                    if (bo(A, C)) {
                        p.rects = {
                            reference: Vl(A, La(C), p.options.strategy === "fixed"),
                            popper: Wn(C)
                        },
                        p.reset = false,
                        p.placement = p.options.placement,
                        p.orderedModifiers.forEach(function($) {
                            return p.modifiersData[$.name] = Object.assign({}, $.data);
                        });
                        for (var S = 0; S < p.orderedModifiers.length; S++) {
                            if (p.reset === true) {
                                p.reset = false,
                                S = -1;
                                continue;
                            }
                            var T = p.orderedModifiers[S]
                              , P = T.fn
                              , _ = T.options
                              , U = _ === void 0 ? {} : _
                              , L = T.name;
                            typeof P == "function" && (p = P({
                                state: p,
                                options: U,
                                name: L,
                                instance: m
                            }) || p);
                        }
                    }
                }
            },
            update: ql(function() {
                return new Promise(function(w) {
                    m.forceUpdate(),
                    w(p);
                }
                );
            }),
            destroy: function() {
                y(),
                f = true;
            }
        };
        if (!bo(s, c))
            return m;
        m.setOptions(d).then(function(w) {
            !f && d.onFirstUpdate && d.onFirstUpdate(w);
        });
        function b() {
            p.orderedModifiers.forEach(function(w) {
                var E = w.name
                  , A = w.options
                  , C = A === void 0 ? {} : A
                  , S = w.effect;
                if (typeof S == "function") {
                    var T = S({
                        state: p,
                        name: E,
                        instance: m,
                        options: C
                    })
                      , P = function() {};
                    u.push(T || P);
                }
            });
        }
        function y() {
            u.forEach(function(w) {
                return w();
            }),
            u = [];
        }
        return m;
    }
    ;
}
var Gl = [xl, Fl, vl, ol, Rl, Pl, $l, dl, jl]
  , Kl = Yl({
    defaultModifiers: Gl
});
function Xl(e) {
    return _t("MuiPopper", e);
}
Rt("MuiPopper", ["root"]);
function Ql(e, t) {
    if (t === "ltr")
        return e;
    switch (e) {
    case "bottom-end":
        return "bottom-start";
    case "bottom-start":
        return "bottom-end";
    case "top-end":
        return "top-start";
    case "top-start":
        return "top-end";
    default:
        return e;
    }
}
function Cn(e) {
    return typeof e == "function" ? e() : e;
}
function Zl(e) {
    return e.nodeType !== void 0;
}
const Jl = (e) => {
    const {classes: t} = e;
    return Dt({
        root: ["root"]
    }, Xl, t);
}
  , es = {}
  , ts = h.forwardRef(function(t, a) {
    const {anchorEl: n, children: o, direction: i, disablePortal: l, modifiers: s, open: c, placement: d, popperOptions: p, popperRef: u, slotProps: f={}, slots: m={}, TransitionProps: b, ownerState: y, ...w} = t
      , E = h.useRef(null)
      , A = xn(E, a)
      , C = h.useRef(null)
      , S = xn(C, u)
      , T = h.useRef(S);
    hn( () => {
        T.current = S;
    }
    , [S]),
    h.useImperativeHandle(u, () => C.current, []);
    const P = Ql(d, i)
      , [_,U] = h.useState(P)
      , [L,$] = h.useState(Cn(n));
    h.useEffect( () => {
        C.current && C.current.forceUpdate();
    }
    ),
    h.useEffect( () => {
        n && $(Cn(n));
    }
    , [n]),
    hn( () => {
        if (!L || !c)
            return;
        const Z = (Ee) => {
            U(Ee.placement);
        }
        ;
        let W = [{
            name: "preventOverflow",
            options: {
                altBoundary: l
            }
        }, {
            name: "flip",
            options: {
                altBoundary: l
            }
        }, {
            name: "onUpdate",
            enabled: true,
            phase: "afterWrite",
            fn: ({state: Ee}) => {
                Z(Ee);
            }
        }];
        s != null && (W = W.concat(s)),
        p && p.modifiers != null && (W = W.concat(p.modifiers));
        const le = Kl(L, E.current, {
            placement: P,
            ...p,
            modifiers: W
        });
        return T.current(le),
        () => {
            le.destroy(),
            T.current(null);
        }
        ;
    }
    , [L, l, s, c, p, P]);
    const q = {
        placement: _
    };
    b !== null && (q.TransitionProps = b);
    const M = Jl(t)
      , G = m.root ?? "div"
      , H = Ea({
        elementType: G,
        externalSlotProps: f.root,
        externalForwardedProps: w,
        additionalProps: {
            role: "tooltip",
            ref: A
        },
        ownerState: t,
        className: M.root
    });
    return r.jsx(G, {
        ...H,
        children: typeof o == "function" ? o(q) : o
    });
})
  , as = h.forwardRef(function(t, a) {
    const {anchorEl: n, children: o, container: i, direction: l="ltr", disablePortal: s=false, keepMounted: c=false, modifiers: d, open: p, placement: u="bottom", popperOptions: f=es, popperRef: m, style: b, transition: y=false, slotProps: w={}, slots: E={}, ...A} = t
      , [C,S] = h.useState(true)
      , T = () => {
        S(false);
    }
      , P = () => {
        S(true);
    }
    ;
    if (!c && !p && (!y || C))
        return null;
    let _;
    if (i)
        _ = i;
    else if (n) {
        const $ = Cn(n);
        _ = $ && Zl($) ? gn($).body : gn(null).body;
    }
    const U = !p && c && (!y || C) ? "none" : void 0
      , L = y ? {
        in: p,
        onEnter: T,
        onExited: P
    } : void 0;
    return r.jsx(Qi, {
        disablePortal: s,
        container: _,
        children: r.jsx(ts, {
            anchorEl: n,
            direction: l,
            disablePortal: s,
            modifiers: d,
            ref: a,
            open: y ? !C : p,
            placement: u,
            popperOptions: f,
            popperRef: m,
            slotProps: w,
            slots: E,
            ...A,
            style: {
                position: "fixed",
                top: 0,
                left: 0,
                display: U,
                ...b
            },
            TransitionProps: L,
            children: o
        })
    });
})
  , ns = xe(as, {
    name: "MuiPopper",
    slot: "Root"
})({})
  , gi = h.forwardRef(function(t, a) {
    const n = Fn()
      , o = Bt({
        props: t,
        name: "MuiPopper"
    })
      , {anchorEl: i, component: l, components: s, componentsProps: c, container: d, disablePortal: p, keepMounted: u, modifiers: f, open: m, placement: b, popperOptions: y, popperRef: w, transition: E, slots: A, slotProps: C, ...S} = o
      , T = (A == null ? void 0 : A.root) ?? (s == null ? void 0 : s.Root)
      , P = {
        anchorEl: i,
        container: d,
        disablePortal: p,
        keepMounted: u,
        modifiers: f,
        open: m,
        placement: b,
        popperOptions: y,
        popperRef: w,
        transition: E,
        ...S
    };
    return r.jsx(ns, {
        as: l,
        direction: n ? "rtl" : "ltr",
        slots: {
            root: T
        },
        slotProps: C ?? c,
        ...P,
        ref: a
    });
});
function os(e) {
    return _t("MuiListSubheader", e);
}
Rt("MuiListSubheader", ["root", "colorPrimary", "colorInherit", "gutters", "inset", "sticky"]);
const is = (e) => {
    const {classes: t, color: a, disableGutters: n, inset: o, disableSticky: i} = e
      , l = {
        root: ["root", a !== "default" && `color${ae(a)}`, !n && "gutters", o && "inset", !i && "sticky"]
    };
    return Dt(l, os, t);
}
  , rs = xe("li", {
    name: "MuiListSubheader",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.root, a.color !== "default" && t[`color${ae(a.color)}`], !a.disableGutters && t.gutters, a.inset && t.inset, !a.disableSticky && t.sticky];
    }
})(it( ({theme: e}) => ({
    boxSizing: "border-box",
    lineHeight: "48px",
    listStyle: "none",
    color: (e.vars || e).palette.text.secondary,
    fontFamily: e.typography.fontFamily,
    fontWeight: e.typography.fontWeightMedium,
    fontSize: e.typography.pxToRem(14),
    variants: [{
        props: {
            color: "primary"
        },
        style: {
            color: (e.vars || e).palette.primary.main
        }
    }, {
        props: {
            color: "inherit"
        },
        style: {
            color: "inherit"
        }
    }, {
        props: ({ownerState: t}) => !t.disableGutters,
        style: {
            paddingLeft: 16,
            paddingRight: 16
        }
    }, {
        props: ({ownerState: t}) => t.inset,
        style: {
            paddingLeft: 72
        }
    }, {
        props: ({ownerState: t}) => !t.disableSticky,
        style: {
            position: "sticky",
            top: 0,
            zIndex: 1,
            backgroundColor: (e.vars || e).palette.background.paper
        }
    }]
})))
  , Sn = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiListSubheader"
    })
      , {className: o, color: i="default", component: l="li", disableGutters: s=false, disableSticky: c=false, inset: d=false, ...p} = n
      , u = {
        ...n,
        color: i,
        component: l,
        disableGutters: s,
        disableSticky: c,
        inset: d
    }
      , f = is(u);
    return r.jsx(rs, {
        as: l,
        className: Se(f.root, o),
        ref: a,
        ownerState: u,
        ...p
    });
});
Sn && (Sn.muiSkipListHighlight = true);
const ls = Ln(r.jsx("path", {
    d: "M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z"
}));
function ss(e) {
    return _t("MuiChip", e);
}
const te = Rt("MuiChip", ["root", "sizeSmall", "sizeMedium", "colorDefault", "colorError", "colorInfo", "colorPrimary", "colorSecondary", "colorSuccess", "colorWarning", "disabled", "clickable", "clickableColorPrimary", "clickableColorSecondary", "deletable", "deletableColorPrimary", "deletableColorSecondary", "outlined", "filled", "outlinedPrimary", "outlinedSecondary", "filledPrimary", "filledSecondary", "avatar", "avatarSmall", "avatarMedium", "avatarColorPrimary", "avatarColorSecondary", "icon", "iconSmall", "iconMedium", "iconColorPrimary", "iconColorSecondary", "label", "labelSmall", "labelMedium", "deleteIcon", "deleteIconSmall", "deleteIconMedium", "deleteIconColorPrimary", "deleteIconColorSecondary", "deleteIconOutlinedColorPrimary", "deleteIconOutlinedColorSecondary", "deleteIconFilledColorPrimary", "deleteIconFilledColorSecondary", "focusVisible"])
  , cs = (e) => {
    const {classes: t, disabled: a, size: n, color: o, iconColor: i, onDelete: l, clickable: s, variant: c} = e
      , d = {
        root: ["root", c, a && "disabled", `size${ae(n)}`, `color${ae(o)}`, s && "clickable", s && `clickableColor${ae(o)}`, l && "deletable", l && `deletableColor${ae(o)}`, `${c}${ae(o)}`],
        label: ["label", `label${ae(n)}`],
        avatar: ["avatar", `avatar${ae(n)}`, `avatarColor${ae(o)}`],
        icon: ["icon", `icon${ae(n)}`, `iconColor${ae(i)}`],
        deleteIcon: ["deleteIcon", `deleteIcon${ae(n)}`, `deleteIconColor${ae(o)}`, `deleteIcon${ae(c)}Color${ae(o)}`]
    };
    return Dt(d, ss, t);
}
  , ps = xe("div", {
    name: "MuiChip",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e
          , {color: n, iconColor: o, clickable: i, onDelete: l, size: s, variant: c} = a;
        return [{
            [`& .${te.avatar}`]: t.avatar
        }, {
            [`& .${te.avatar}`]: t[`avatar${ae(s)}`]
        }, {
            [`& .${te.avatar}`]: t[`avatarColor${ae(n)}`]
        }, {
            [`& .${te.icon}`]: t.icon
        }, {
            [`& .${te.icon}`]: t[`icon${ae(s)}`]
        }, {
            [`& .${te.icon}`]: t[`iconColor${ae(o)}`]
        }, {
            [`& .${te.deleteIcon}`]: t.deleteIcon
        }, {
            [`& .${te.deleteIcon}`]: t[`deleteIcon${ae(s)}`]
        }, {
            [`& .${te.deleteIcon}`]: t[`deleteIconColor${ae(n)}`]
        }, {
            [`& .${te.deleteIcon}`]: t[`deleteIcon${ae(c)}Color${ae(n)}`]
        }, t.root, t[`size${ae(s)}`], t[`color${ae(n)}`], i && t.clickable, i && n !== "default" && t[`clickableColor${ae(n)})`], l && t.deletable, l && n !== "default" && t[`deletableColor${ae(n)}`], t[c], t[`${c}${ae(n)}`]];
    }
})(it( ({theme: e}) => {
    const t = e.palette.mode === "light" ? e.palette.grey[700] : e.palette.grey[300];
    return {
        maxWidth: "100%",
        fontFamily: e.typography.fontFamily,
        fontSize: e.typography.pxToRem(13),
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        height: 32,
        color: (e.vars || e).palette.text.primary,
        backgroundColor: (e.vars || e).palette.action.selected,
        borderRadius: 32 / 2,
        whiteSpace: "nowrap",
        transition: e.transitions.create(["background-color", "box-shadow"]),
        cursor: "unset",
        outline: 0,
        textDecoration: "none",
        border: 0,
        padding: 0,
        verticalAlign: "middle",
        boxSizing: "border-box",
        [`&.${te.disabled}`]: {
            opacity: (e.vars || e).palette.action.disabledOpacity,
            pointerEvents: "none"
        },
        [`& .${te.avatar}`]: {
            marginLeft: 5,
            marginRight: -6,
            width: 24,
            height: 24,
            color: e.vars ? e.vars.palette.Chip.defaultAvatarColor : t,
            fontSize: e.typography.pxToRem(12)
        },
        [`& .${te.avatarColorPrimary}`]: {
            color: (e.vars || e).palette.primary.contrastText,
            backgroundColor: (e.vars || e).palette.primary.dark
        },
        [`& .${te.avatarColorSecondary}`]: {
            color: (e.vars || e).palette.secondary.contrastText,
            backgroundColor: (e.vars || e).palette.secondary.dark
        },
        [`& .${te.avatarSmall}`]: {
            marginLeft: 4,
            marginRight: -4,
            width: 18,
            height: 18,
            fontSize: e.typography.pxToRem(10)
        },
        [`& .${te.icon}`]: {
            marginLeft: 5,
            marginRight: -6
        },
        [`& .${te.deleteIcon}`]: {
            WebkitTapHighlightColor: "transparent",
            color: e.vars ? `rgba(${e.vars.palette.text.primaryChannel} / 0.26)` : tt(e.palette.text.primary, 0.26),
            fontSize: 22,
            cursor: "pointer",
            margin: "0 5px 0 -6px",
            "&:hover": {
                color: e.vars ? `rgba(${e.vars.palette.text.primaryChannel} / 0.4)` : tt(e.palette.text.primary, 0.4)
            }
        },
        variants: [{
            props: {
                size: "small"
            },
            style: {
                height: 24,
                [`& .${te.icon}`]: {
                    fontSize: 18,
                    marginLeft: 4,
                    marginRight: -4
                },
                [`& .${te.deleteIcon}`]: {
                    fontSize: 16,
                    marginRight: 4,
                    marginLeft: -4
                }
            }
        }, ...Object.entries(e.palette).filter(Ba(["contrastText"])).map( ([a]) => ({
            props: {
                color: a
            },
            style: {
                backgroundColor: (e.vars || e).palette[a].main,
                color: (e.vars || e).palette[a].contrastText,
                [`& .${te.deleteIcon}`]: {
                    color: e.vars ? `rgba(${e.vars.palette[a].contrastTextChannel} / 0.7)` : tt(e.palette[a].contrastText, 0.7),
                    "&:hover, &:active": {
                        color: (e.vars || e).palette[a].contrastText
                    }
                }
            }
        })), {
            props: (a) => a.iconColor === a.color,
            style: {
                [`& .${te.icon}`]: {
                    color: e.vars ? e.vars.palette.Chip.defaultIconColor : t
                }
            }
        }, {
            props: (a) => a.iconColor === a.color && a.color !== "default",
            style: {
                [`& .${te.icon}`]: {
                    color: "inherit"
                }
            }
        }, {
            props: {
                onDelete: true
            },
            style: {
                [`&.${te.focusVisible}`]: {
                    backgroundColor: e.vars ? `rgba(${e.vars.palette.action.selectedChannel} / calc(${e.vars.palette.action.selectedOpacity} + ${e.vars.palette.action.focusOpacity}))` : tt(e.palette.action.selected, e.palette.action.selectedOpacity + e.palette.action.focusOpacity)
                }
            }
        }, ...Object.entries(e.palette).filter(Ba(["dark"])).map( ([a]) => ({
            props: {
                color: a,
                onDelete: true
            },
            style: {
                [`&.${te.focusVisible}`]: {
                    background: (e.vars || e).palette[a].dark
                }
            }
        })), {
            props: {
                clickable: true
            },
            style: {
                userSelect: "none",
                WebkitTapHighlightColor: "transparent",
                cursor: "pointer",
                "&:hover": {
                    backgroundColor: e.vars ? `rgba(${e.vars.palette.action.selectedChannel} / calc(${e.vars.palette.action.selectedOpacity} + ${e.vars.palette.action.hoverOpacity}))` : tt(e.palette.action.selected, e.palette.action.selectedOpacity + e.palette.action.hoverOpacity)
                },
                [`&.${te.focusVisible}`]: {
                    backgroundColor: e.vars ? `rgba(${e.vars.palette.action.selectedChannel} / calc(${e.vars.palette.action.selectedOpacity} + ${e.vars.palette.action.focusOpacity}))` : tt(e.palette.action.selected, e.palette.action.selectedOpacity + e.palette.action.focusOpacity)
                },
                "&:active": {
                    boxShadow: (e.vars || e).shadows[1]
                }
            }
        }, ...Object.entries(e.palette).filter(Ba(["dark"])).map( ([a]) => ({
            props: {
                color: a,
                clickable: true
            },
            style: {
                [`&:hover, &.${te.focusVisible}`]: {
                    backgroundColor: (e.vars || e).palette[a].dark
                }
            }
        })), {
            props: {
                variant: "outlined"
            },
            style: {
                backgroundColor: "transparent",
                border: e.vars ? `1px solid ${e.vars.palette.Chip.defaultBorder}` : `1px solid ${e.palette.mode === "light" ? e.palette.grey[400] : e.palette.grey[700]}`,
                [`&.${te.clickable}:hover`]: {
                    backgroundColor: (e.vars || e).palette.action.hover
                },
                [`&.${te.focusVisible}`]: {
                    backgroundColor: (e.vars || e).palette.action.focus
                },
                [`& .${te.avatar}`]: {
                    marginLeft: 4
                },
                [`& .${te.avatarSmall}`]: {
                    marginLeft: 2
                },
                [`& .${te.icon}`]: {
                    marginLeft: 4
                },
                [`& .${te.iconSmall}`]: {
                    marginLeft: 2
                },
                [`& .${te.deleteIcon}`]: {
                    marginRight: 5
                },
                [`& .${te.deleteIconSmall}`]: {
                    marginRight: 3
                }
            }
        }, ...Object.entries(e.palette).filter(Ba()).map( ([a]) => ({
            props: {
                variant: "outlined",
                color: a
            },
            style: {
                color: (e.vars || e).palette[a].main,
                border: `1px solid ${e.vars ? `rgba(${e.vars.palette[a].mainChannel} / 0.7)` : tt(e.palette[a].main, 0.7)}`,
                [`&.${te.clickable}:hover`]: {
                    backgroundColor: e.vars ? `rgba(${e.vars.palette[a].mainChannel} / ${e.vars.palette.action.hoverOpacity})` : tt(e.palette[a].main, e.palette.action.hoverOpacity)
                },
                [`&.${te.focusVisible}`]: {
                    backgroundColor: e.vars ? `rgba(${e.vars.palette[a].mainChannel} / ${e.vars.palette.action.focusOpacity})` : tt(e.palette[a].main, e.palette.action.focusOpacity)
                },
                [`& .${te.deleteIcon}`]: {
                    color: e.vars ? `rgba(${e.vars.palette[a].mainChannel} / 0.7)` : tt(e.palette[a].main, 0.7),
                    "&:hover, &:active": {
                        color: (e.vars || e).palette[a].main
                    }
                }
            }
        }))]
    };
}
))
  , ds = xe("span", {
    name: "MuiChip",
    slot: "Label",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e
          , {size: n} = a;
        return [t.label, t[`label${ae(n)}`]];
    }
})({
    overflow: "hidden",
    textOverflow: "ellipsis",
    paddingLeft: 12,
    paddingRight: 12,
    whiteSpace: "nowrap",
    variants: [{
        props: {
            variant: "outlined"
        },
        style: {
            paddingLeft: 11,
            paddingRight: 11
        }
    }, {
        props: {
            size: "small"
        },
        style: {
            paddingLeft: 8,
            paddingRight: 8
        }
    }, {
        props: {
            size: "small",
            variant: "outlined"
        },
        style: {
            paddingLeft: 7,
            paddingRight: 7
        }
    }]
});
function yo(e) {
    return e.key === "Backspace" || e.key === "Delete";
}
const Qn = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiChip"
    })
      , {avatar: o, className: i, clickable: l, color: s="default", component: c, deleteIcon: d, disabled: p=false, icon: u, label: f, onClick: m, onDelete: b, onKeyDown: y, onKeyUp: w, size: E="medium", variant: A="filled", tabIndex: C, skipFocusWhenDisabled: S=false, slots: T={}, slotProps: P={}, ..._} = n
      , U = h.useRef(null)
      , L = xn(U, a)
      , $ = (D) => {
        D.stopPropagation(),
        b && b(D);
    }
      , q = (D) => {
        D.currentTarget === D.target && yo(D) && D.preventDefault(),
        y && y(D);
    }
      , M = (D) => {
        D.currentTarget === D.target && b && yo(D) && b(D),
        w && w(D);
    }
      , G = l !== false && m ? true : l
      , H = G || b ? Aa : c || "div"
      , Z = {
        ...n,
        component: H,
        disabled: p,
        size: E,
        color: s,
        iconColor: h.isValidElement(u) && u.props.color || s,
        onDelete: !!b,
        clickable: G,
        variant: A
    }
      , W = cs(Z)
      , le = H === Aa ? {
        component: c || "div",
        focusVisibleClassName: W.focusVisible,
        ...b && {
            disableRipple: true
        }
    } : {};
    let Ee = null;
    b && (Ee = d && h.isValidElement(d) ? h.cloneElement(d, {
        className: Se(d.props.className, W.deleteIcon),
        onClick: $
    }) : r.jsx(ls, {
        className: W.deleteIcon,
        onClick: $
    }));
    let ne = null;
    o && h.isValidElement(o) && (ne = h.cloneElement(o, {
        className: Se(W.avatar, o.props.className)
    }));
    let Ae = null;
    u && h.isValidElement(u) && (Ae = h.cloneElement(u, {
        className: Se(W.icon, u.props.className)
    }));
    const pe = {
        slots: T,
        slotProps: P
    }
      , [se,Ge] = at("root", {
        elementType: ps,
        externalForwardedProps: {
            ...pe,
            ..._
        },
        ownerState: Z,
        shouldForwardComponentProp: true,
        ref: L,
        className: Se(W.root, i),
        additionalProps: {
            disabled: G && p ? true : void 0,
            tabIndex: S && p ? -1 : C,
            ...le
        },
        getSlotProps: (D) => ({
            ...D,
            onClick: (he) => {
                var _a2;
                (_a2 = D.onClick) == null ? void 0 : _a2.call(D, he),
                m(he);
            }
            ,
            onKeyDown: (he) => {
                var _a2;
                (_a2 = D.onKeyDown) == null ? void 0 : _a2.call(D, he),
                q(he);
            }
            ,
            onKeyUp: (he) => {
                var _a2;
                (_a2 = D.onKeyUp) == null ? void 0 : _a2.call(D, he),
                M(he);
            }
        })
    })
      , [ye,ce] = at("label", {
        elementType: ds,
        externalForwardedProps: pe,
        ownerState: Z,
        className: W.label
    });
    return r.jsxs(se, {
        as: H,
        ...Ge,
        children: [ne || Ae, r.jsx(ye, {
            ...ce,
            children: f
        }), Ee]
    });
});
function us(e) {
    return _t("MuiAutocomplete", e);
}
const X = Rt("MuiAutocomplete", ["root", "expanded", "fullWidth", "focused", "focusVisible", "tag", "tagSizeSmall", "tagSizeMedium", "hasPopupIcon", "hasClearIcon", "inputRoot", "input", "inputFocused", "endAdornment", "clearIndicator", "popupIndicator", "popupIndicatorOpen", "popper", "popperDisablePortal", "paper", "listbox", "loading", "noOptions", "option", "groupLabel", "groupUl"]);
var wo, Co;
const ms = (e) => {
    const {classes: t, disablePortal: a, expanded: n, focused: o, fullWidth: i, hasClearIcon: l, hasPopupIcon: s, inputFocused: c, popupOpen: d, size: p} = e
      , u = {
        root: ["root", n && "expanded", o && "focused", i && "fullWidth", l && "hasClearIcon", s && "hasPopupIcon"],
        inputRoot: ["inputRoot"],
        input: ["input", c && "inputFocused"],
        tag: ["tag", `tagSize${ae(p)}`],
        endAdornment: ["endAdornment"],
        clearIndicator: ["clearIndicator"],
        popupIndicator: ["popupIndicator", d && "popupIndicatorOpen"],
        popper: ["popper", a && "popperDisablePortal"],
        paper: ["paper"],
        listbox: ["listbox"],
        loading: ["loading"],
        noOptions: ["noOptions"],
        option: ["option"],
        groupLabel: ["groupLabel"],
        groupUl: ["groupUl"]
    };
    return Dt(u, us, t);
}
  , fs = xe("div", {
    name: "MuiAutocomplete",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e
          , {fullWidth: n, hasClearIcon: o, hasPopupIcon: i, inputFocused: l, size: s} = a;
        return [{
            [`& .${X.tag}`]: t.tag
        }, {
            [`& .${X.tag}`]: t[`tagSize${ae(s)}`]
        }, {
            [`& .${X.inputRoot}`]: t.inputRoot
        }, {
            [`& .${X.input}`]: t.input
        }, {
            [`& .${X.input}`]: l && t.inputFocused
        }, t.root, n && t.fullWidth, i && t.hasPopupIcon, o && t.hasClearIcon];
    }
})({
    [`&.${X.focused} .${X.clearIndicator}`]: {
        visibility: "visible"
    },
    "@media (pointer: fine)": {
        [`&:hover .${X.clearIndicator}`]: {
            visibility: "visible"
        }
    },
    [`& .${X.tag}`]: {
        margin: 3,
        maxWidth: "calc(100% - 6px)"
    },
    [`& .${X.inputRoot}`]: {
        [`.${X.hasPopupIcon}&, .${X.hasClearIcon}&`]: {
            paddingRight: 30
        },
        [`.${X.hasPopupIcon}.${X.hasClearIcon}&`]: {
            paddingRight: 56
        },
        [`& .${X.input}`]: {
            width: 0,
            minWidth: 30
        }
    },
    [`& .${pn.root}`]: {
        paddingBottom: 1,
        "& .MuiInput-input": {
            padding: "4px 4px 4px 0px"
        }
    },
    [`& .${pn.root}.${Gt.sizeSmall}`]: {
        [`& .${pn.input}`]: {
            padding: "2px 4px 3px 0"
        }
    },
    [`& .${io.root}`]: {
        padding: 9,
        [`.${X.hasPopupIcon}&, .${X.hasClearIcon}&`]: {
            paddingRight: 39
        },
        [`.${X.hasPopupIcon}.${X.hasClearIcon}&`]: {
            paddingRight: 65
        },
        [`& .${X.input}`]: {
            padding: "7.5px 4px 7.5px 5px"
        },
        [`& .${X.endAdornment}`]: {
            right: 9
        }
    },
    [`& .${io.root}.${Gt.sizeSmall}`]: {
        paddingTop: 6,
        paddingBottom: 6,
        paddingLeft: 6,
        [`& .${X.input}`]: {
            padding: "2.5px 4px 2.5px 8px"
        }
    },
    [`& .${la.root}`]: {
        paddingTop: 19,
        paddingLeft: 8,
        [`.${X.hasPopupIcon}&, .${X.hasClearIcon}&`]: {
            paddingRight: 39
        },
        [`.${X.hasPopupIcon}.${X.hasClearIcon}&`]: {
            paddingRight: 65
        },
        [`& .${la.input}`]: {
            padding: "7px 4px"
        },
        [`& .${X.endAdornment}`]: {
            right: 9
        }
    },
    [`& .${la.root}.${Gt.sizeSmall}`]: {
        paddingBottom: 1,
        [`& .${la.input}`]: {
            padding: "2.5px 4px"
        }
    },
    [`& .${Gt.hiddenLabel}`]: {
        paddingTop: 8
    },
    [`& .${la.root}.${Gt.hiddenLabel}`]: {
        paddingTop: 0,
        paddingBottom: 0,
        [`& .${X.input}`]: {
            paddingTop: 16,
            paddingBottom: 17
        }
    },
    [`& .${la.root}.${Gt.hiddenLabel}.${Gt.sizeSmall}`]: {
        [`& .${X.input}`]: {
            paddingTop: 8,
            paddingBottom: 9
        }
    },
    [`& .${X.input}`]: {
        flexGrow: 1,
        textOverflow: "ellipsis",
        opacity: 0
    },
    variants: [{
        props: {
            fullWidth: true
        },
        style: {
            width: "100%"
        }
    }, {
        props: {
            size: "small"
        },
        style: {
            [`& .${X.tag}`]: {
                margin: 2,
                maxWidth: "calc(100% - 4px)"
            }
        }
    }, {
        props: {
            inputFocused: true
        },
        style: {
            [`& .${X.input}`]: {
                opacity: 1
            }
        }
    }, {
        props: {
            multiple: true
        },
        style: {
            [`& .${X.inputRoot}`]: {
                flexWrap: "wrap"
            }
        }
    }]
})
  , vs = xe("div", {
    name: "MuiAutocomplete",
    slot: "EndAdornment"
})({
    position: "absolute",
    right: 0,
    top: "50%",
    transform: "translate(0, -50%)"
})
  , gs = xe(Mn, {
    name: "MuiAutocomplete",
    slot: "ClearIndicator"
})({
    marginRight: -2,
    padding: 4,
    visibility: "hidden"
})
  , xs = xe(Mn, {
    name: "MuiAutocomplete",
    slot: "PopupIndicator",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.popupIndicator, a.popupOpen && t.popupIndicatorOpen];
    }
})({
    padding: 2,
    marginRight: -2,
    variants: [{
        props: {
            popupOpen: true
        },
        style: {
            transform: "rotate(180deg)"
        }
    }]
})
  , hs = xe(gi, {
    name: "MuiAutocomplete",
    slot: "Popper",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [{
            [`& .${X.option}`]: t.option
        }, t.popper, a.disablePortal && t.popperDisablePortal];
    }
})(it( ({theme: e}) => ({
    zIndex: (e.vars || e).zIndex.modal,
    variants: [{
        props: {
            disablePortal: true
        },
        style: {
            position: "absolute"
        }
    }]
})))
  , bs = xe(Yo, {
    name: "MuiAutocomplete",
    slot: "Paper"
})(it( ({theme: e}) => ({
    ...e.typography.body1,
    overflow: "auto"
})))
  , ys = xe("div", {
    name: "MuiAutocomplete",
    slot: "Loading"
})(it( ({theme: e}) => ({
    color: (e.vars || e).palette.text.secondary,
    padding: "14px 16px"
})))
  , ws = xe("div", {
    name: "MuiAutocomplete",
    slot: "NoOptions"
})(it( ({theme: e}) => ({
    color: (e.vars || e).palette.text.secondary,
    padding: "14px 16px"
})))
  , Cs = xe("ul", {
    name: "MuiAutocomplete",
    slot: "Listbox"
})(it( ({theme: e}) => ({
    listStyle: "none",
    margin: 0,
    padding: "8px 0",
    maxHeight: "40vh",
    overflow: "auto",
    position: "relative",
    [`& .${X.option}`]: {
        minHeight: 48,
        display: "flex",
        overflow: "hidden",
        justifyContent: "flex-start",
        alignItems: "center",
        cursor: "pointer",
        paddingTop: 6,
        boxSizing: "border-box",
        outline: "0",
        WebkitTapHighlightColor: "transparent",
        paddingBottom: 6,
        paddingLeft: 16,
        paddingRight: 16,
        [e.breakpoints.up("sm")]: {
            minHeight: "auto"
        },
        [`&.${X.focused}`]: {
            backgroundColor: (e.vars || e).palette.action.hover,
            "@media (hover: none)": {
                backgroundColor: "transparent"
            }
        },
        '&[aria-disabled="true"]': {
            opacity: (e.vars || e).palette.action.disabledOpacity,
            pointerEvents: "none"
        },
        [`&.${X.focusVisible}`]: {
            backgroundColor: (e.vars || e).palette.action.focus
        },
        '&[aria-selected="true"]': {
            backgroundColor: e.vars ? `rgba(${e.vars.palette.primary.mainChannel} / ${e.vars.palette.action.selectedOpacity})` : tt(e.palette.primary.main, e.palette.action.selectedOpacity),
            [`&.${X.focused}`]: {
                backgroundColor: e.vars ? `rgba(${e.vars.palette.primary.mainChannel} / calc(${e.vars.palette.action.selectedOpacity} + ${e.vars.palette.action.hoverOpacity}))` : tt(e.palette.primary.main, e.palette.action.selectedOpacity + e.palette.action.hoverOpacity),
                "@media (hover: none)": {
                    backgroundColor: (e.vars || e).palette.action.selected
                }
            },
            [`&.${X.focusVisible}`]: {
                backgroundColor: e.vars ? `rgba(${e.vars.palette.primary.mainChannel} / calc(${e.vars.palette.action.selectedOpacity} + ${e.vars.palette.action.focusOpacity}))` : tt(e.palette.primary.main, e.palette.action.selectedOpacity + e.palette.action.focusOpacity)
            }
        }
    }
})))
  , Ss = xe(Sn, {
    name: "MuiAutocomplete",
    slot: "GroupLabel"
})(it( ({theme: e}) => ({
    backgroundColor: (e.vars || e).palette.background.paper,
    top: -8
})))
  , Es = xe("ul", {
    name: "MuiAutocomplete",
    slot: "GroupUl"
})({
    padding: 0,
    [`& .${X.option}`]: {
        paddingLeft: 24
    }
})
  , As = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiAutocomplete"
    })
      , {autoComplete: o=false, autoHighlight: i=false, autoSelect: l=false, blurOnSelect: s=false, ChipProps: c, className: d, clearIcon: p=wo || (wo = r.jsx(Zi, {
        fontSize: "small"
    })), clearOnBlur: u=!n.freeSolo, clearOnEscape: f=false, clearText: m="Clear", closeText: b="Close", componentsProps: y, defaultValue: w=n.multiple ? [] : null, disableClearable: E=false, disableCloseOnSelect: A=false, disabled: C=false, disabledItemsFocusable: S=false, disableListWrap: T=false, disablePortal: P=false, filterOptions: _, filterSelectedOptions: U=false, forcePopupIcon: L="auto", freeSolo: $=false, fullWidth: q=false, getLimitTagsText: M= (ie) => `+${ie}`, getOptionDisabled: G, getOptionKey: H, getOptionLabel: Z, isOptionEqualToValue: W, groupBy: le, handleHomeEndKeys: Ee=!n.freeSolo, id: ne, includeInputInList: Ae=false, inputValue: pe, limitTags: se=-1, ListboxComponent: Ge, ListboxProps: ye, loading: ce=false, loadingText: D="Loading\u2026", multiple: he=false, noOptionsText: Le="No options", onChange: ue, onClose: ke, onHighlightChange: qe, onInputChange: Qe, onOpen: Oe, open: rt, openOnFocus: St=false, openText: me="Open", options: Te, PaperComponent: R, PopperComponent: Me, popupIcon: Y=Co || (Co = r.jsx(Ji, {})), readOnly: _e=false, renderGroup: Ke, renderInput: lt, renderOption: k, renderTags: I, renderValue: oe, selectOnFocus: fe=!n.freeSolo, size: ve="medium", slots: st={}, slotProps: we={}, value: re, ...Be} = n
      , {getRootProps: Et, getInputProps: $e, getInputLabelProps: ct, getPopupIndicatorProps: Xe, getClearProps: He, getItemProps: At, getListboxProps: pt, getOptionProps: ga, value: We, dirty: yt, expanded: wt, id: Vt, popupOpen: dt, focused: ia, focusedItem: ra, anchorEl: Mt, setAnchorEl: O, inputValue: B, groupedOptions: J} = Hr({
        ...n,
        componentName: "Autocomplete"
    })
      , ee = !E && !C && yt && !_e
      , ge = (!$ || L === true) && L !== false
      , {onMouseDown: Ze} = $e()
      , {ref: Ct, ...kt} = pt()
      , Tt = Z || ( (ie) => ie.label ?? ie)
      , Ne = {
        ...n,
        disablePortal: P,
        expanded: wt,
        focused: ia,
        fullWidth: q,
        getOptionLabel: Tt,
        hasClearIcon: ee,
        hasPopupIcon: ge,
        inputFocused: ra === -1,
        popupOpen: dt,
        size: ve
    }
      , ze = ms(Ne)
      , Ht = {
        slots: {
            paper: R,
            popper: Me,
            ...st
        },
        slotProps: {
            chip: c,
            listbox: ye,
            ...y,
            ...we
        }
    }
      , [ha,$a] = at("listbox", {
        elementType: Cs,
        externalForwardedProps: Ht,
        ownerState: Ne,
        className: ze.listbox,
        additionalProps: kt,
        ref: Ct
    })
      , [x,v] = at("paper", {
        elementType: Yo,
        externalForwardedProps: Ht,
        ownerState: Ne,
        className: ze.paper
    })
      , [j,Q] = at("popper", {
        elementType: gi,
        externalForwardedProps: Ht,
        ownerState: Ne,
        className: ze.popper,
        additionalProps: {
            disablePortal: P,
            style: {
                width: Mt ? Mt.clientWidth : null
            },
            role: "presentation",
            anchorEl: Mt,
            open: dt
        }
    });
    let V;
    const Pe = (ie) => ({
        className: ze.tag,
        disabled: C,
        ...At(ie)
    });
    if (he ? We.length > 0 && (I ? V = I(We, Pe, Ne) : oe ? V = oe(We, Pe, Ne) : V = We.map( (ie, Ut) => {
        const {key: qt, ...ba} = Pe({
            index: Ut
        });
        return r.jsx(Qn, {
            label: Tt(ie),
            size: ve,
            ...ba,
            ...Ht.slotProps.chip
        }, qt);
    }
    )) : oe && We != null && (V = oe(We, Pe, Ne)),
    se > -1 && Array.isArray(V)) {
        const ie = V.length - se;
        !ia && ie > 0 && (V = V.splice(0, se),
        V.push(r.jsx("span", {
            className: ze.tag,
            children: M(ie)
        }, V.length)));
    }
    const ut = Ke || ( (ie) => r.jsxs("li", {
        children: [r.jsx(Ss, {
            className: ze.groupLabel,
            ownerState: Ne,
            component: "div",
            children: ie.group
        }), r.jsx(Es, {
            className: ze.groupUl,
            ownerState: Ne,
            children: ie.children
        })]
    }, ie.key))
      , qi = k || ( (ie, Ut) => {
        const {key: qt, ...ba} = ie;
        return r.jsx("li", {
            ...ba,
            children: Tt(Ut)
        }, qt);
    }
    )
      , ao = (ie, Ut) => {
        const qt = ga({
            option: ie,
            index: Ut
        });
        return qi({
            ...qt,
            className: ze.option
        }, ie, {
            selected: qt["aria-selected"],
            index: Ut,
            inputValue: B
        }, Ne);
    }
      , no = Ht.slotProps.clearIndicator
      , oo = Ht.slotProps.popupIndicator;
    return r.jsxs(h.Fragment, {
        children: [r.jsx(fs, {
            ref: a,
            className: Se(ze.root, d),
            ownerState: Ne,
            ...Et(Be),
            children: lt({
                id: Vt,
                disabled: C,
                fullWidth: true,
                size: ve === "small" ? "small" : void 0,
                InputLabelProps: ct(),
                InputProps: {
                    ref: O,
                    className: ze.inputRoot,
                    startAdornment: V,
                    onMouseDown: (ie) => {
                        ie.target === ie.currentTarget && Ze(ie);
                    }
                    ,
                    ...(ee || ge) && {
                        endAdornment: r.jsxs(vs, {
                            className: ze.endAdornment,
                            ownerState: Ne,
                            children: [ee ? r.jsx(gs, {
                                ...He(),
                                "aria-label": m,
                                title: m,
                                ownerState: Ne,
                                ...no,
                                className: Se(ze.clearIndicator, no == null ? void 0 : no.className),
                                children: p
                            }) : null, ge ? r.jsx(xs, {
                                ...Xe(),
                                disabled: C,
                                "aria-label": dt ? b : me,
                                title: dt ? b : me,
                                ownerState: Ne,
                                ...oo,
                                className: Se(ze.popupIndicator, oo == null ? void 0 : oo.className),
                                children: Y
                            }) : null]
                        })
                    }
                },
                inputProps: {
                    className: ze.input,
                    disabled: C,
                    readOnly: _e,
                    ...$e()
                }
            })
        }), Mt ? r.jsx(hs, {
            as: j,
            ...Q,
            children: r.jsxs(bs, {
                as: x,
                ...v,
                children: [ce && J.length === 0 ? r.jsx(ys, {
                    className: ze.loading,
                    ownerState: Ne,
                    children: D
                }) : null, J.length === 0 && !$ && !ce ? r.jsx(ws, {
                    className: ze.noOptions,
                    ownerState: Ne,
                    role: "presentation",
                    onMouseDown: (ie) => {
                        ie.preventDefault();
                    }
                    ,
                    children: Le
                }) : null, J.length > 0 ? r.jsx(ha, {
                    as: Ge,
                    ...$a,
                    children: J.map( (ie, Ut) => le ? ut({
                        key: ie.key,
                        group: ie.group,
                        children: ie.options.map( (qt, ba) => ao(qt, ie.index + ba))
                    }) : ao(ie, Ut))
                }) : null]
            })
        }) : null]
    });
});
function ks(e) {
    return _t("MuiCardActionArea", e);
}
const dn = Rt("MuiCardActionArea", ["root", "focusVisible", "focusHighlight"])
  , Ts = (e) => {
    const {classes: t} = e;
    return Dt({
        root: ["root"],
        focusHighlight: ["focusHighlight"]
    }, ks, t);
}
  , Ps = xe(Aa, {
    name: "MuiCardActionArea",
    slot: "Root"
})(it( ({theme: e}) => ({
    display: "block",
    textAlign: "inherit",
    borderRadius: "inherit",
    width: "100%",
    [`&:hover .${dn.focusHighlight}`]: {
        opacity: (e.vars || e).palette.action.hoverOpacity,
        "@media (hover: none)": {
            opacity: 0
        }
    },
    [`&.${dn.focusVisible} .${dn.focusHighlight}`]: {
        opacity: (e.vars || e).palette.action.focusOpacity
    }
})))
  , Os = xe("span", {
    name: "MuiCardActionArea",
    slot: "FocusHighlight"
})(it( ({theme: e}) => ({
    overflow: "hidden",
    pointerEvents: "none",
    position: "absolute",
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    borderRadius: "inherit",
    opacity: 0,
    backgroundColor: "currentcolor",
    transition: e.transitions.create("opacity", {
        duration: e.transitions.duration.short
    })
})))
  , js = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiCardActionArea"
    })
      , {children: o, className: i, focusVisibleClassName: l, slots: s={}, slotProps: c={}, ...d} = n
      , p = n
      , u = Ts(p)
      , f = {
        slots: s,
        slotProps: c
    }
      , [m,b] = at("root", {
        elementType: Ps,
        externalForwardedProps: {
            ...f,
            ...d
        },
        shouldForwardComponentProp: true,
        ownerState: p,
        ref: a,
        className: Se(u.root, i),
        additionalProps: {
            focusVisibleClassName: Se(l, u.focusVisible)
        }
    })
      , [y,w] = at("focusHighlight", {
        elementType: Os,
        externalForwardedProps: f,
        ownerState: p,
        ref: a,
        className: u.focusHighlight
    });
    return r.jsxs(m, {
        ...b,
        children: [o, r.jsx(y, {
            ...w
        })]
    });
});
function Is(e) {
    return _t("MuiCardActions", e);
}
Rt("MuiCardActions", ["root", "spacing"]);
const _s = (e) => {
    const {classes: t, disableSpacing: a} = e;
    return Dt({
        root: ["root", !a && "spacing"]
    }, Is, t);
}
  , Rs = xe("div", {
    name: "MuiCardActions",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.root, !a.disableSpacing && t.spacing];
    }
})({
    display: "flex",
    alignItems: "center",
    padding: 8,
    variants: [{
        props: {
            disableSpacing: false
        },
        style: {
            "& > :not(style) ~ :not(style)": {
                marginLeft: 8
            }
        }
    }]
})
  , Zn = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiCardActions"
    })
      , {disableSpacing: o=false, className: i, ...l} = n
      , s = {
        ...n,
        disableSpacing: o
    }
      , c = _s(s);
    return r.jsx(Rs, {
        className: Se(c.root, i),
        ownerState: s,
        ref: a,
        ...l
    });
});
function Ds(e) {
    return _t("MuiTab", e);
}
const mt = Rt("MuiTab", ["root", "labelIcon", "textColorInherit", "textColorPrimary", "textColorSecondary", "selected", "disabled", "fullWidth", "wrapped", "iconWrapper", "icon"])
  , Fs = (e) => {
    const {classes: t, textColor: a, fullWidth: n, wrapped: o, icon: i, label: l, selected: s, disabled: c} = e
      , d = {
        root: ["root", i && l && "labelIcon", `textColor${ae(a)}`, n && "fullWidth", o && "wrapped", s && "selected", c && "disabled"],
        icon: ["iconWrapper", "icon"]
    };
    return Dt(d, Ds, t);
}
  , Ls = xe(Aa, {
    name: "MuiTab",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.root, a.label && a.icon && t.labelIcon, t[`textColor${ae(a.textColor)}`], a.fullWidth && t.fullWidth, a.wrapped && t.wrapped, {
            [`& .${mt.iconWrapper}`]: t.iconWrapper
        }, {
            [`& .${mt.icon}`]: t.icon
        }];
    }
})(it( ({theme: e}) => ({
    ...e.typography.button,
    maxWidth: 360,
    minWidth: 90,
    position: "relative",
    minHeight: 48,
    flexShrink: 0,
    padding: "12px 16px",
    overflow: "hidden",
    whiteSpace: "normal",
    textAlign: "center",
    lineHeight: 1.25,
    variants: [{
        props: ({ownerState: t}) => t.label && (t.iconPosition === "top" || t.iconPosition === "bottom"),
        style: {
            flexDirection: "column"
        }
    }, {
        props: ({ownerState: t}) => t.label && t.iconPosition !== "top" && t.iconPosition !== "bottom",
        style: {
            flexDirection: "row"
        }
    }, {
        props: ({ownerState: t}) => t.icon && t.label,
        style: {
            minHeight: 72,
            paddingTop: 9,
            paddingBottom: 9
        }
    }, {
        props: ({ownerState: t, iconPosition: a}) => t.icon && t.label && a === "top",
        style: {
            [`& > .${mt.icon}`]: {
                marginBottom: 6
            }
        }
    }, {
        props: ({ownerState: t, iconPosition: a}) => t.icon && t.label && a === "bottom",
        style: {
            [`& > .${mt.icon}`]: {
                marginTop: 6
            }
        }
    }, {
        props: ({ownerState: t, iconPosition: a}) => t.icon && t.label && a === "start",
        style: {
            [`& > .${mt.icon}`]: {
                marginRight: e.spacing(1)
            }
        }
    }, {
        props: ({ownerState: t, iconPosition: a}) => t.icon && t.label && a === "end",
        style: {
            [`& > .${mt.icon}`]: {
                marginLeft: e.spacing(1)
            }
        }
    }, {
        props: {
            textColor: "inherit"
        },
        style: {
            color: "inherit",
            opacity: 0.6,
            [`&.${mt.selected}`]: {
                opacity: 1
            },
            [`&.${mt.disabled}`]: {
                opacity: (e.vars || e).palette.action.disabledOpacity
            }
        }
    }, {
        props: {
            textColor: "primary"
        },
        style: {
            color: (e.vars || e).palette.text.secondary,
            [`&.${mt.selected}`]: {
                color: (e.vars || e).palette.primary.main
            },
            [`&.${mt.disabled}`]: {
                color: (e.vars || e).palette.text.disabled
            }
        }
    }, {
        props: {
            textColor: "secondary"
        },
        style: {
            color: (e.vars || e).palette.text.secondary,
            [`&.${mt.selected}`]: {
                color: (e.vars || e).palette.secondary.main
            },
            [`&.${mt.disabled}`]: {
                color: (e.vars || e).palette.text.disabled
            }
        }
    }, {
        props: ({ownerState: t}) => t.fullWidth,
        style: {
            flexShrink: 1,
            flexGrow: 1,
            flexBasis: 0,
            maxWidth: "none"
        }
    }, {
        props: ({ownerState: t}) => t.wrapped,
        style: {
            fontSize: e.typography.pxToRem(12)
        }
    }]
})))
  , Ms = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiTab"
    })
      , {className: o, disabled: i=false, disableFocusRipple: l=false, fullWidth: s, icon: c, iconPosition: d="top", indicator: p, label: u, onChange: f, onClick: m, onFocus: b, selected: y, selectionFollowsFocus: w, textColor: E="inherit", value: A, wrapped: C=false, ...S} = n
      , T = {
        ...n,
        disabled: i,
        disableFocusRipple: l,
        selected: y,
        icon: !!c,
        iconPosition: d,
        label: !!u,
        fullWidth: s,
        textColor: E,
        wrapped: C
    }
      , P = Fs(T)
      , _ = c && u && h.isValidElement(c) ? h.cloneElement(c, {
        className: Se(P.icon, c.props.className)
    }) : c
      , U = ($) => {
        !y && f && f($, A),
        m && m($);
    }
      , L = ($) => {
        w && !y && f && f($, A),
        b && b($);
    }
    ;
    return r.jsxs(Ls, {
        focusRipple: !l,
        className: Se(P.root, o),
        ref: a,
        role: "tab",
        "aria-selected": y,
        disabled: i,
        onClick: U,
        onFocus: L,
        ownerState: T,
        tabIndex: y ? 0 : -1,
        ...S,
        children: [d === "top" || d === "start" ? r.jsxs(h.Fragment, {
            children: [_, u]
        }) : r.jsxs(h.Fragment, {
            children: [u, _]
        }), p]
    });
})
  , $s = Ln(r.jsx("path", {
    d: "M15.41 16.09l-4.58-4.59 4.58-4.59L14 5.5l-6 6 6 6z"
}))
  , Bs = Ln(r.jsx("path", {
    d: "M8.59 16.34l4.58-4.59-4.58-4.59L10 5.75l6 6-6 6z"
}));
function Ns(e) {
    return (1 + Math.sin(Math.PI * e - Math.PI / 2)) / 2;
}
function zs(e, t, a, n={}, o= () => {}
) {
    const {ease: i=Ns, duration: l=300} = n;
    let s = null;
    const c = t[e];
    let d = false;
    const p = () => {
        d = true;
    }
      , u = (f) => {
        if (d) {
            o(new Error("Animation cancelled"));
            return;
        }
        s === null && (s = f);
        const m = Math.min(1, (f - s) / l);
        if (t[e] = i(m) * (a - c) + c,
        m >= 1) {
            requestAnimationFrame( () => {
                o(null);
            }
            );
            return;
        }
        requestAnimationFrame(u);
    }
    ;
    return c === a ? (o(new Error("Element already at target position")),
    p) : (requestAnimationFrame(u),
    p);
}
const Vs = {
    width: 99,
    height: 99,
    position: "absolute",
    top: -9999,
    overflow: "scroll"
};
function Hs(e) {
    const {onChange: t, ...a} = e
      , n = h.useRef()
      , o = h.useRef(null)
      , i = () => {
        n.current = o.current.offsetHeight - o.current.clientHeight;
    }
    ;
    return hn( () => {
        const l = Go( () => {
            const c = n.current;
            i(),
            c !== n.current && t(n.current);
        }
        )
          , s = Ko(o.current);
        return s.addEventListener("resize", l),
        () => {
            l.clear(),
            s.removeEventListener("resize", l);
        }
        ;
    }
    , [t]),
    h.useEffect( () => {
        i(),
        t(n.current);
    }
    , [t]),
    r.jsx("div", {
        style: Vs,
        ...a,
        ref: o
    });
}
function Us(e) {
    return _t("MuiTabScrollButton", e);
}
const qs = Rt("MuiTabScrollButton", ["root", "vertical", "horizontal", "disabled"])
  , Ws = (e) => {
    const {classes: t, orientation: a, disabled: n} = e;
    return Dt({
        root: ["root", a, n && "disabled"]
    }, Us, t);
}
  , Ys = xe(Aa, {
    name: "MuiTabScrollButton",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.root, a.orientation && t[a.orientation]];
    }
})({
    width: 40,
    flexShrink: 0,
    opacity: 0.8,
    [`&.${qs.disabled}`]: {
        opacity: 0
    },
    variants: [{
        props: {
            orientation: "vertical"
        },
        style: {
            width: "100%",
            height: 40,
            "& svg": {
                transform: "var(--TabScrollButton-svgRotate)"
            }
        }
    }]
})
  , Gs = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiTabScrollButton"
    })
      , {className: o, slots: i={}, slotProps: l={}, direction: s, orientation: c, disabled: d, ...p} = n
      , u = Fn()
      , f = {
        isRtl: u,
        ...n
    }
      , m = Ws(f)
      , b = i.StartScrollButtonIcon ?? $s
      , y = i.EndScrollButtonIcon ?? Bs
      , w = Ea({
        elementType: b,
        externalSlotProps: l.startScrollButtonIcon,
        additionalProps: {
            fontSize: "small"
        },
        ownerState: f
    })
      , E = Ea({
        elementType: y,
        externalSlotProps: l.endScrollButtonIcon,
        additionalProps: {
            fontSize: "small"
        },
        ownerState: f
    });
    return r.jsx(Ys, {
        component: "div",
        className: Se(m.root, o),
        ref: a,
        role: null,
        ownerState: f,
        tabIndex: null,
        ...p,
        style: {
            ...p.style,
            ...c === "vertical" && {
                "--TabScrollButton-svgRotate": `rotate(${u ? -90 : 90}deg)`
            }
        },
        children: s === "left" ? r.jsx(b, {
            ...w
        }) : r.jsx(y, {
            ...E
        })
    });
});
function Ks(e) {
    return _t("MuiTabs", e);
}
const un = Rt("MuiTabs", ["root", "vertical", "list", "flexContainer", "flexContainerVertical", "centered", "scroller", "fixed", "scrollableX", "scrollableY", "hideScrollbar", "scrollButtons", "scrollButtonsHideMobile", "indicator"])
  , So = (e, t) => e === t ? e.firstChild : t && t.nextElementSibling ? t.nextElementSibling : e.firstChild
  , Eo = (e, t) => e === t ? e.lastChild : t && t.previousElementSibling ? t.previousElementSibling : e.lastChild
  , Ha = (e, t, a) => {
    let n = false
      , o = a(e, t);
    for (; o; ) {
        if (o === e.firstChild) {
            if (n)
                return;
            n = true;
        }
        const i = o.disabled || o.getAttribute("aria-disabled") === "true";
        if (!o.hasAttribute("tabindex") || i)
            o = a(e, o);
        else {
            o.focus();
            return;
        }
    }
}
  , Xs = (e) => {
    const {vertical: t, fixed: a, hideScrollbar: n, scrollableX: o, scrollableY: i, centered: l, scrollButtonsHideMobile: s, classes: c} = e;
    return Dt({
        root: ["root", t && "vertical"],
        scroller: ["scroller", a && "fixed", n && "hideScrollbar", o && "scrollableX", i && "scrollableY"],
        list: ["list", "flexContainer", t && "flexContainerVertical", t && "vertical", l && "centered"],
        indicator: ["indicator"],
        scrollButtons: ["scrollButtons", s && "scrollButtonsHideMobile"],
        scrollableX: [o && "scrollableX"],
        hideScrollbar: [n && "hideScrollbar"]
    }, Ks, c);
}
  , Qs = xe("div", {
    name: "MuiTabs",
    slot: "Root",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [{
            [`& .${un.scrollButtons}`]: t.scrollButtons
        }, {
            [`& .${un.scrollButtons}`]: a.scrollButtonsHideMobile && t.scrollButtonsHideMobile
        }, t.root, a.vertical && t.vertical];
    }
})(it( ({theme: e}) => ({
    overflow: "hidden",
    minHeight: 48,
    WebkitOverflowScrolling: "touch",
    display: "flex",
    variants: [{
        props: ({ownerState: t}) => t.vertical,
        style: {
            flexDirection: "column"
        }
    }, {
        props: ({ownerState: t}) => t.scrollButtonsHideMobile,
        style: {
            [`& .${un.scrollButtons}`]: {
                [e.breakpoints.down("sm")]: {
                    display: "none"
                }
            }
        }
    }]
})))
  , Zs = xe("div", {
    name: "MuiTabs",
    slot: "Scroller",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.scroller, a.fixed && t.fixed, a.hideScrollbar && t.hideScrollbar, a.scrollableX && t.scrollableX, a.scrollableY && t.scrollableY];
    }
})({
    position: "relative",
    display: "inline-block",
    flex: "1 1 auto",
    whiteSpace: "nowrap",
    variants: [{
        props: ({ownerState: e}) => e.fixed,
        style: {
            overflowX: "hidden",
            width: "100%"
        }
    }, {
        props: ({ownerState: e}) => e.hideScrollbar,
        style: {
            scrollbarWidth: "none",
            "&::-webkit-scrollbar": {
                display: "none"
            }
        }
    }, {
        props: ({ownerState: e}) => e.scrollableX,
        style: {
            overflowX: "auto",
            overflowY: "hidden"
        }
    }, {
        props: ({ownerState: e}) => e.scrollableY,
        style: {
            overflowY: "auto",
            overflowX: "hidden"
        }
    }]
})
  , Js = xe("div", {
    name: "MuiTabs",
    slot: "List",
    overridesResolver: (e, t) => {
        const {ownerState: a} = e;
        return [t.list, t.flexContainer, a.vertical && t.flexContainerVertical, a.centered && t.centered];
    }
})({
    display: "flex",
    variants: [{
        props: ({ownerState: e}) => e.vertical,
        style: {
            flexDirection: "column"
        }
    }, {
        props: ({ownerState: e}) => e.centered,
        style: {
            justifyContent: "center"
        }
    }]
})
  , ec = xe("span", {
    name: "MuiTabs",
    slot: "Indicator"
})(it( ({theme: e}) => ({
    position: "absolute",
    height: 2,
    bottom: 0,
    width: "100%",
    transition: e.transitions.create(),
    variants: [{
        props: {
            indicatorColor: "primary"
        },
        style: {
            backgroundColor: (e.vars || e).palette.primary.main
        }
    }, {
        props: {
            indicatorColor: "secondary"
        },
        style: {
            backgroundColor: (e.vars || e).palette.secondary.main
        }
    }, {
        props: ({ownerState: t}) => t.vertical,
        style: {
            height: "100%",
            width: 2,
            right: 0
        }
    }]
})))
  , tc = xe(Hs)({
    overflowX: "auto",
    overflowY: "hidden",
    scrollbarWidth: "none",
    "&::-webkit-scrollbar": {
        display: "none"
    }
})
  , Ao = {}
  , ac = h.forwardRef(function(t, a) {
    const n = Bt({
        props: t,
        name: "MuiTabs"
    })
      , o = er()
      , i = Fn()
      , {"aria-label": l, "aria-labelledby": s, action: c, centered: d=false, children: p, className: u, component: f="div", allowScrollButtonsMobile: m=false, indicatorColor: b="primary", onChange: y, orientation: w="horizontal", ScrollButtonComponent: E, scrollButtons: A="auto", selectionFollowsFocus: C, slots: S={}, slotProps: T={}, TabIndicatorProps: P={}, TabScrollButtonProps: _={}, textColor: U="primary", value: L, variant: $="standard", visibleScrollbar: q=false, ...M} = n
      , G = $ === "scrollable"
      , H = w === "vertical"
      , Z = H ? "scrollTop" : "scrollLeft"
      , W = H ? "top" : "left"
      , le = H ? "bottom" : "right"
      , Ee = H ? "clientHeight" : "clientWidth"
      , ne = H ? "height" : "width"
      , Ae = {
        ...n,
        component: f,
        allowScrollButtonsMobile: m,
        indicatorColor: b,
        orientation: w,
        vertical: H,
        scrollButtons: A,
        textColor: U,
        variant: $,
        visibleScrollbar: q,
        fixed: !G,
        hideScrollbar: G && !q,
        scrollableX: G && !H,
        scrollableY: G && H,
        centered: d && !G,
        scrollButtonsHideMobile: !m
    }
      , pe = Xs(Ae)
      , se = Ea({
        elementType: S.StartScrollButtonIcon,
        externalSlotProps: T.startScrollButtonIcon,
        ownerState: Ae
    })
      , Ge = Ea({
        elementType: S.EndScrollButtonIcon,
        externalSlotProps: T.endScrollButtonIcon,
        ownerState: Ae
    })
      , [ye,ce] = h.useState(false)
      , [D,he] = h.useState(Ao)
      , [Le,ue] = h.useState(false)
      , [ke,qe] = h.useState(false)
      , [Qe,Oe] = h.useState(false)
      , [rt,St] = h.useState({
        overflow: "hidden",
        scrollbarWidth: 0
    })
      , me = /* @__PURE__ */
    new Map()
      , Te = h.useRef(null)
      , R = h.useRef(null)
      , Me = {
        slots: S,
        slotProps: {
            indicator: P,
            scrollButton: _,
            ...T
        }
    }
      , Y = () => {
        const O = Te.current;
        let B;
        if (O) {
            const ee = O.getBoundingClientRect();
            B = {
                clientWidth: O.clientWidth,
                scrollLeft: O.scrollLeft,
                scrollTop: O.scrollTop,
                scrollWidth: O.scrollWidth,
                top: ee.top,
                bottom: ee.bottom,
                left: ee.left,
                right: ee.right
            };
        }
        let J;
        if (O && L !== false) {
            const ee = R.current.children;
            if (ee.length > 0) {
                const ge = ee[me.get(L)];
                J = ge ? ge.getBoundingClientRect() : null;
            }
        }
        return {
            tabsMeta: B,
            tabMeta: J
        };
    }
      , _e = Qt( () => {
        const {tabsMeta: O, tabMeta: B} = Y();
        let J = 0, ee;
        H ? (ee = "top",
        B && O && (J = B.top - O.top + O.scrollTop)) : (ee = i ? "right" : "left",
        B && O && (J = (i ? -1 : 1) * (B[ee] - O[ee] + O.scrollLeft)));
        const ge = {
            [ee]: J,
            [ne]: B ? B[ne] : 0
        };
        if (typeof D[ee] != "number" || typeof D[ne] != "number")
            he(ge);
        else {
            const Ze = Math.abs(D[ee] - ge[ee])
              , Ct = Math.abs(D[ne] - ge[ne]);
            (Ze >= 1 || Ct >= 1) && he(ge);
        }
    }
    )
      , Ke = (O, {animation: B=true}={}) => {
        B ? zs(Z, Te.current, O, {
            duration: o.transitions.duration.standard
        }) : Te.current[Z] = O;
    }
      , lt = (O) => {
        let B = Te.current[Z];
        H ? B += O : B += O * (i ? -1 : 1),
        Ke(B);
    }
      , k = () => {
        const O = Te.current[Ee];
        let B = 0;
        const J = Array.from(R.current.children);
        for (let ee = 0; ee < J.length; ee += 1) {
            const ge = J[ee];
            if (B + ge[Ee] > O) {
                ee === 0 && (B = O);
                break;
            }
            B += ge[Ee];
        }
        return B;
    }
      , I = () => {
        lt(-1 * k());
    }
      , oe = () => {
        lt(k());
    }
      , [fe,{onChange: ve, ...st}] = at("scrollbar", {
        className: Se(pe.scrollableX, pe.hideScrollbar),
        elementType: tc,
        shouldForwardComponentProp: true,
        externalForwardedProps: Me,
        ownerState: Ae
    })
      , we = h.useCallback( (O) => {
        ve == null ? void 0 : ve(O),
        St({
            overflow: null,
            scrollbarWidth: O
        });
    }
    , [ve])
      , [re,Be] = at("scrollButtons", {
        className: Se(pe.scrollButtons, _.className),
        elementType: Gs,
        externalForwardedProps: Me,
        ownerState: Ae,
        additionalProps: {
            orientation: w,
            slots: {
                StartScrollButtonIcon: S.startScrollButtonIcon || S.StartScrollButtonIcon,
                EndScrollButtonIcon: S.endScrollButtonIcon || S.EndScrollButtonIcon
            },
            slotProps: {
                startScrollButtonIcon: se,
                endScrollButtonIcon: Ge
            }
        }
    })
      , Et = () => {
        const O = {};
        O.scrollbarSizeListener = G ? r.jsx(fe, {
            ...st,
            onChange: we
        }) : null;
        const J = G && (A === "auto" && (Le || ke) || A === true);
        return O.scrollButtonStart = J ? r.jsx(re, {
            direction: i ? "right" : "left",
            onClick: I,
            disabled: !Le,
            ...Be
        }) : null,
        O.scrollButtonEnd = J ? r.jsx(re, {
            direction: i ? "left" : "right",
            onClick: oe,
            disabled: !ke,
            ...Be
        }) : null,
        O;
    }
      , $e = Qt( (O) => {
        const {tabsMeta: B, tabMeta: J} = Y();
        if (!(!J || !B)) {
            if (J[W] < B[W]) {
                const ee = B[Z] + (J[W] - B[W]);
                Ke(ee, {
                    animation: O
                });
            } else if (J[le] > B[le]) {
                const ee = B[Z] + (J[le] - B[le]);
                Ke(ee, {
                    animation: O
                });
            }
        }
    }
    )
      , ct = Qt( () => {
        G && A !== false && Oe(!Qe);
    }
    );
    h.useEffect( () => {
        const O = Go( () => {
            Te.current && _e();
        }
        );
        let B;
        const J = (Ze) => {
            Ze.forEach( (Ct) => {
                Ct.removedNodes.forEach( (kt) => {
                    B == null ? void 0 : B.unobserve(kt);
                }
                ),
                Ct.addedNodes.forEach( (kt) => {
                    B == null ? void 0 : B.observe(kt);
                }
                );
            }
            ),
            O(),
            ct();
        }
          , ee = Ko(Te.current);
        ee.addEventListener("resize", O);
        let ge;
        return typeof ResizeObserver < "u" && (B = new ResizeObserver(O),
        Array.from(R.current.children).forEach( (Ze) => {
            B.observe(Ze);
        }
        )),
        typeof MutationObserver < "u" && (ge = new MutationObserver(J),
        ge.observe(R.current, {
            childList: true
        })),
        () => {
            O.clear(),
            ee.removeEventListener("resize", O),
            ge == null ? void 0 : ge.disconnect(),
            B == null ? void 0 : B.disconnect();
        }
        ;
    }
    , [_e, ct]),
    h.useEffect( () => {
        const O = Array.from(R.current.children)
          , B = O.length;
        if (typeof IntersectionObserver < "u" && B > 0 && G && A !== false) {
            const J = O[0]
              , ee = O[B - 1]
              , ge = {
                root: Te.current,
                threshold: 0.99
            }
              , Ze = (Tt) => {
                ue(!Tt[0].isIntersecting);
            }
              , Ct = new IntersectionObserver(Ze,ge);
            Ct.observe(J);
            const kt = (Tt) => {
                qe(!Tt[0].isIntersecting);
            }
              , xa = new IntersectionObserver(kt,ge);
            return xa.observe(ee),
            () => {
                Ct.disconnect(),
                xa.disconnect();
            }
            ;
        }
    }
    , [G, A, Qe, p == null ? void 0 : p.length]),
    h.useEffect( () => {
        ce(true);
    }
    , []),
    h.useEffect( () => {
        _e();
    }
    ),
    h.useEffect( () => {
        $e(Ao !== D);
    }
    , [$e, D]),
    h.useImperativeHandle(c, () => ({
        updateIndicator: _e,
        updateScrollButtons: ct
    }), [_e, ct]);
    const [Xe,He] = at("indicator", {
        className: Se(pe.indicator, P.className),
        elementType: ec,
        externalForwardedProps: Me,
        ownerState: Ae,
        additionalProps: {
            style: D
        }
    })
      , At = r.jsx(Xe, {
        ...He
    });
    let pt = 0;
    const ga = h.Children.map(p, (O) => {
        if (!h.isValidElement(O))
            return null;
        const B = O.props.value === void 0 ? pt : O.props.value;
        me.set(B, pt);
        const J = B === L;
        return pt += 1,
        h.cloneElement(O, {
            fullWidth: $ === "fullWidth",
            indicator: J && !ye && At,
            selected: J,
            selectionFollowsFocus: C,
            onChange: y,
            textColor: U,
            value: B,
            ...pt === 1 && L === false && !O.props.tabIndex ? {
                tabIndex: 0
            } : {}
        });
    }
    )
      , We = (O) => {
        if (O.altKey || O.shiftKey || O.ctrlKey || O.metaKey)
            return;
        const B = R.current
          , J = gn(B).activeElement;
        if (J.getAttribute("role") !== "tab")
            return;
        let ge = w === "horizontal" ? "ArrowLeft" : "ArrowUp"
          , Ze = w === "horizontal" ? "ArrowRight" : "ArrowDown";
        switch (w === "horizontal" && i && (ge = "ArrowRight",
        Ze = "ArrowLeft"),
        O.key) {
        case ge:
            O.preventDefault(),
            Ha(B, J, Eo);
            break;
        case Ze:
            O.preventDefault(),
            Ha(B, J, So);
            break;
        case "Home":
            O.preventDefault(),
            Ha(B, null, So);
            break;
        case "End":
            O.preventDefault(),
            Ha(B, null, Eo);
            break;
        }
    }
      , yt = Et()
      , [wt,Vt] = at("root", {
        ref: a,
        className: Se(pe.root, u),
        elementType: Qs,
        externalForwardedProps: {
            ...Me,
            ...M,
            component: f
        },
        ownerState: Ae
    })
      , [dt,ia] = at("scroller", {
        ref: Te,
        className: pe.scroller,
        elementType: Zs,
        externalForwardedProps: Me,
        ownerState: Ae,
        additionalProps: {
            style: {
                overflow: rt.overflow,
                [H ? `margin${i ? "Left" : "Right"}` : "marginBottom"]: q ? void 0 : -rt.scrollbarWidth
            }
        }
    })
      , [ra,Mt] = at("list", {
        ref: R,
        className: Se(pe.list, pe.flexContainer),
        elementType: Js,
        externalForwardedProps: Me,
        ownerState: Ae,
        getSlotProps: (O) => ({
            ...O,
            onKeyDown: (B) => {
                var _a2;
                We(B),
                (_a2 = O.onKeyDown) == null ? void 0 : _a2.call(O, B);
            }
        })
    });
    return r.jsxs(wt, {
        ...Vt,
        children: [yt.scrollButtonStart, yt.scrollbarSizeListener, r.jsxs(dt, {
            ...ia,
            children: [r.jsx(ra, {
                "aria-label": l,
                "aria-labelledby": s,
                "aria-orientation": w === "vertical" ? "vertical" : null,
                role: "tablist",
                ...Mt,
                children: ga
            }), ye && At]
        }), yt.scrollButtonEnd]
    });
})
  , va = () => (Qa.setLanguage("ru-RU"),
Qa)
  , xi = (e="0,0.00") => {
    const t = va();
    return (a) => t(a.value).format(e);
}
  , hi = () => {
    const e = va();
    return (t) => e(t.value).format("0.00%");
}
  , nc = (e) => {
    let t = "";
    return e >= 1048576 ? t = `${e / 1048576} MB` : e >= 1024 ? t = `${e / 1024} KB` : t = `${e} B`,
    t;
}
  , oc = (e) => e === "audio/*" || e === "video/*" || e === "image/*" || e === "text/*" || /^(application|audio|font|example|image|message|model|multipart|text|video)\/[\w.+-]+$/.test(e)
  , ic = (e) => /^.*\.\w+$/.test(e)
  , rc = (e) => tr(e) ? Object.entries(e).filter( ([t]) => oc(t)).map( ([t,a]) => {
    const n = a.filter( (o) => ic(o));
    return n.length > 0 ? [t, ...n].join(",") : t;
}
).join(",") : "";
function na(e, t, a, n) {
    function o(i) {
        return i instanceof a ? i : new a(function(l) {
            l(i);
        }
        );
    }
    return new (a || (a = Promise))(function(i, l) {
        function s(p) {
            try {
                d(n.next(p));
            } catch (u) {
                l(u);
            }
        }
        function c(p) {
            try {
                d(n.throw(p));
            } catch (u) {
                l(u);
            }
        }
        function d(p) {
            p.done ? i(p.value) : o(p.value).then(s, c);
        }
        d((n = n.apply(e, t || [])).next());
    }
    );
}
const lc = /* @__PURE__ */
new Map([["1km", "application/vnd.1000minds.decision-model+xml"], ["3dml", "text/vnd.in3d.3dml"], ["3ds", "image/x-3ds"], ["3g2", "video/3gpp2"], ["3gp", "video/3gp"], ["3gpp", "video/3gpp"], ["3mf", "model/3mf"], ["7z", "application/x-7z-compressed"], ["7zip", "application/x-7z-compressed"], ["123", "application/vnd.lotus-1-2-3"], ["aab", "application/x-authorware-bin"], ["aac", "audio/x-acc"], ["aam", "application/x-authorware-map"], ["aas", "application/x-authorware-seg"], ["abw", "application/x-abiword"], ["ac", "application/vnd.nokia.n-gage.ac+xml"], ["ac3", "audio/ac3"], ["acc", "application/vnd.americandynamics.acc"], ["ace", "application/x-ace-compressed"], ["acu", "application/vnd.acucobol"], ["acutc", "application/vnd.acucorp"], ["adp", "audio/adpcm"], ["aep", "application/vnd.audiograph"], ["afm", "application/x-font-type1"], ["afp", "application/vnd.ibm.modcap"], ["ahead", "application/vnd.ahead.space"], ["ai", "application/pdf"], ["aif", "audio/x-aiff"], ["aifc", "audio/x-aiff"], ["aiff", "audio/x-aiff"], ["air", "application/vnd.adobe.air-application-installer-package+zip"], ["ait", "application/vnd.dvb.ait"], ["ami", "application/vnd.amiga.ami"], ["amr", "audio/amr"], ["apk", "application/vnd.android.package-archive"], ["apng", "image/apng"], ["appcache", "text/cache-manifest"], ["application", "application/x-ms-application"], ["apr", "application/vnd.lotus-approach"], ["arc", "application/x-freearc"], ["arj", "application/x-arj"], ["asc", "application/pgp-signature"], ["asf", "video/x-ms-asf"], ["asm", "text/x-asm"], ["aso", "application/vnd.accpac.simply.aso"], ["asx", "video/x-ms-asf"], ["atc", "application/vnd.acucorp"], ["atom", "application/atom+xml"], ["atomcat", "application/atomcat+xml"], ["atomdeleted", "application/atomdeleted+xml"], ["atomsvc", "application/atomsvc+xml"], ["atx", "application/vnd.antix.game-component"], ["au", "audio/x-au"], ["avi", "video/x-msvideo"], ["avif", "image/avif"], ["aw", "application/applixware"], ["azf", "application/vnd.airzip.filesecure.azf"], ["azs", "application/vnd.airzip.filesecure.azs"], ["azv", "image/vnd.airzip.accelerator.azv"], ["azw", "application/vnd.amazon.ebook"], ["b16", "image/vnd.pco.b16"], ["bat", "application/x-msdownload"], ["bcpio", "application/x-bcpio"], ["bdf", "application/x-font-bdf"], ["bdm", "application/vnd.syncml.dm+wbxml"], ["bdoc", "application/x-bdoc"], ["bed", "application/vnd.realvnc.bed"], ["bh2", "application/vnd.fujitsu.oasysprs"], ["bin", "application/octet-stream"], ["blb", "application/x-blorb"], ["blorb", "application/x-blorb"], ["bmi", "application/vnd.bmi"], ["bmml", "application/vnd.balsamiq.bmml+xml"], ["bmp", "image/bmp"], ["book", "application/vnd.framemaker"], ["box", "application/vnd.previewsystems.box"], ["boz", "application/x-bzip2"], ["bpk", "application/octet-stream"], ["bpmn", "application/octet-stream"], ["bsp", "model/vnd.valve.source.compiled-map"], ["btif", "image/prs.btif"], ["buffer", "application/octet-stream"], ["bz", "application/x-bzip"], ["bz2", "application/x-bzip2"], ["c", "text/x-c"], ["c4d", "application/vnd.clonk.c4group"], ["c4f", "application/vnd.clonk.c4group"], ["c4g", "application/vnd.clonk.c4group"], ["c4p", "application/vnd.clonk.c4group"], ["c4u", "application/vnd.clonk.c4group"], ["c11amc", "application/vnd.cluetrust.cartomobile-config"], ["c11amz", "application/vnd.cluetrust.cartomobile-config-pkg"], ["cab", "application/vnd.ms-cab-compressed"], ["caf", "audio/x-caf"], ["cap", "application/vnd.tcpdump.pcap"], ["car", "application/vnd.curl.car"], ["cat", "application/vnd.ms-pki.seccat"], ["cb7", "application/x-cbr"], ["cba", "application/x-cbr"], ["cbr", "application/x-cbr"], ["cbt", "application/x-cbr"], ["cbz", "application/x-cbr"], ["cc", "text/x-c"], ["cco", "application/x-cocoa"], ["cct", "application/x-director"], ["ccxml", "application/ccxml+xml"], ["cdbcmsg", "application/vnd.contact.cmsg"], ["cda", "application/x-cdf"], ["cdf", "application/x-netcdf"], ["cdfx", "application/cdfx+xml"], ["cdkey", "application/vnd.mediastation.cdkey"], ["cdmia", "application/cdmi-capability"], ["cdmic", "application/cdmi-container"], ["cdmid", "application/cdmi-domain"], ["cdmio", "application/cdmi-object"], ["cdmiq", "application/cdmi-queue"], ["cdr", "application/cdr"], ["cdx", "chemical/x-cdx"], ["cdxml", "application/vnd.chemdraw+xml"], ["cdy", "application/vnd.cinderella"], ["cer", "application/pkix-cert"], ["cfs", "application/x-cfs-compressed"], ["cgm", "image/cgm"], ["chat", "application/x-chat"], ["chm", "application/vnd.ms-htmlhelp"], ["chrt", "application/vnd.kde.kchart"], ["cif", "chemical/x-cif"], ["cii", "application/vnd.anser-web-certificate-issue-initiation"], ["cil", "application/vnd.ms-artgalry"], ["cjs", "application/node"], ["cla", "application/vnd.claymore"], ["class", "application/octet-stream"], ["clkk", "application/vnd.crick.clicker.keyboard"], ["clkp", "application/vnd.crick.clicker.palette"], ["clkt", "application/vnd.crick.clicker.template"], ["clkw", "application/vnd.crick.clicker.wordbank"], ["clkx", "application/vnd.crick.clicker"], ["clp", "application/x-msclip"], ["cmc", "application/vnd.cosmocaller"], ["cmdf", "chemical/x-cmdf"], ["cml", "chemical/x-cml"], ["cmp", "application/vnd.yellowriver-custom-menu"], ["cmx", "image/x-cmx"], ["cod", "application/vnd.rim.cod"], ["coffee", "text/coffeescript"], ["com", "application/x-msdownload"], ["conf", "text/plain"], ["cpio", "application/x-cpio"], ["cpp", "text/x-c"], ["cpt", "application/mac-compactpro"], ["crd", "application/x-mscardfile"], ["crl", "application/pkix-crl"], ["crt", "application/x-x509-ca-cert"], ["crx", "application/x-chrome-extension"], ["cryptonote", "application/vnd.rig.cryptonote"], ["csh", "application/x-csh"], ["csl", "application/vnd.citationstyles.style+xml"], ["csml", "chemical/x-csml"], ["csp", "application/vnd.commonspace"], ["csr", "application/octet-stream"], ["css", "text/css"], ["cst", "application/x-director"], ["csv", "text/csv"], ["cu", "application/cu-seeme"], ["curl", "text/vnd.curl"], ["cww", "application/prs.cww"], ["cxt", "application/x-director"], ["cxx", "text/x-c"], ["dae", "model/vnd.collada+xml"], ["daf", "application/vnd.mobius.daf"], ["dart", "application/vnd.dart"], ["dataless", "application/vnd.fdsn.seed"], ["davmount", "application/davmount+xml"], ["dbf", "application/vnd.dbf"], ["dbk", "application/docbook+xml"], ["dcr", "application/x-director"], ["dcurl", "text/vnd.curl.dcurl"], ["dd2", "application/vnd.oma.dd2+xml"], ["ddd", "application/vnd.fujixerox.ddd"], ["ddf", "application/vnd.syncml.dmddf+xml"], ["dds", "image/vnd.ms-dds"], ["deb", "application/x-debian-package"], ["def", "text/plain"], ["deploy", "application/octet-stream"], ["der", "application/x-x509-ca-cert"], ["dfac", "application/vnd.dreamfactory"], ["dgc", "application/x-dgc-compressed"], ["dic", "text/x-c"], ["dir", "application/x-director"], ["dis", "application/vnd.mobius.dis"], ["disposition-notification", "message/disposition-notification"], ["dist", "application/octet-stream"], ["distz", "application/octet-stream"], ["djv", "image/vnd.djvu"], ["djvu", "image/vnd.djvu"], ["dll", "application/octet-stream"], ["dmg", "application/x-apple-diskimage"], ["dmn", "application/octet-stream"], ["dmp", "application/vnd.tcpdump.pcap"], ["dms", "application/octet-stream"], ["dna", "application/vnd.dna"], ["doc", "application/msword"], ["docm", "application/vnd.ms-word.template.macroEnabled.12"], ["docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"], ["dot", "application/msword"], ["dotm", "application/vnd.ms-word.template.macroEnabled.12"], ["dotx", "application/vnd.openxmlformats-officedocument.wordprocessingml.template"], ["dp", "application/vnd.osgi.dp"], ["dpg", "application/vnd.dpgraph"], ["dra", "audio/vnd.dra"], ["drle", "image/dicom-rle"], ["dsc", "text/prs.lines.tag"], ["dssc", "application/dssc+der"], ["dtb", "application/x-dtbook+xml"], ["dtd", "application/xml-dtd"], ["dts", "audio/vnd.dts"], ["dtshd", "audio/vnd.dts.hd"], ["dump", "application/octet-stream"], ["dvb", "video/vnd.dvb.file"], ["dvi", "application/x-dvi"], ["dwd", "application/atsc-dwd+xml"], ["dwf", "model/vnd.dwf"], ["dwg", "image/vnd.dwg"], ["dxf", "image/vnd.dxf"], ["dxp", "application/vnd.spotfire.dxp"], ["dxr", "application/x-director"], ["ear", "application/java-archive"], ["ecelp4800", "audio/vnd.nuera.ecelp4800"], ["ecelp7470", "audio/vnd.nuera.ecelp7470"], ["ecelp9600", "audio/vnd.nuera.ecelp9600"], ["ecma", "application/ecmascript"], ["edm", "application/vnd.novadigm.edm"], ["edx", "application/vnd.novadigm.edx"], ["efif", "application/vnd.picsel"], ["ei6", "application/vnd.pg.osasli"], ["elc", "application/octet-stream"], ["emf", "image/emf"], ["eml", "message/rfc822"], ["emma", "application/emma+xml"], ["emotionml", "application/emotionml+xml"], ["emz", "application/x-msmetafile"], ["eol", "audio/vnd.digital-winds"], ["eot", "application/vnd.ms-fontobject"], ["eps", "application/postscript"], ["epub", "application/epub+zip"], ["es", "application/ecmascript"], ["es3", "application/vnd.eszigno3+xml"], ["esa", "application/vnd.osgi.subsystem"], ["esf", "application/vnd.epson.esf"], ["et3", "application/vnd.eszigno3+xml"], ["etx", "text/x-setext"], ["eva", "application/x-eva"], ["evy", "application/x-envoy"], ["exe", "application/octet-stream"], ["exi", "application/exi"], ["exp", "application/express"], ["exr", "image/aces"], ["ext", "application/vnd.novadigm.ext"], ["ez", "application/andrew-inset"], ["ez2", "application/vnd.ezpix-album"], ["ez3", "application/vnd.ezpix-package"], ["f", "text/x-fortran"], ["f4v", "video/mp4"], ["f77", "text/x-fortran"], ["f90", "text/x-fortran"], ["fbs", "image/vnd.fastbidsheet"], ["fcdt", "application/vnd.adobe.formscentral.fcdt"], ["fcs", "application/vnd.isac.fcs"], ["fdf", "application/vnd.fdf"], ["fdt", "application/fdt+xml"], ["fe_launch", "application/vnd.denovo.fcselayout-link"], ["fg5", "application/vnd.fujitsu.oasysgp"], ["fgd", "application/x-director"], ["fh", "image/x-freehand"], ["fh4", "image/x-freehand"], ["fh5", "image/x-freehand"], ["fh7", "image/x-freehand"], ["fhc", "image/x-freehand"], ["fig", "application/x-xfig"], ["fits", "image/fits"], ["flac", "audio/x-flac"], ["fli", "video/x-fli"], ["flo", "application/vnd.micrografx.flo"], ["flv", "video/x-flv"], ["flw", "application/vnd.kde.kivio"], ["flx", "text/vnd.fmi.flexstor"], ["fly", "text/vnd.fly"], ["fm", "application/vnd.framemaker"], ["fnc", "application/vnd.frogans.fnc"], ["fo", "application/vnd.software602.filler.form+xml"], ["for", "text/x-fortran"], ["fpx", "image/vnd.fpx"], ["frame", "application/vnd.framemaker"], ["fsc", "application/vnd.fsc.weblaunch"], ["fst", "image/vnd.fst"], ["ftc", "application/vnd.fluxtime.clip"], ["fti", "application/vnd.anser-web-funds-transfer-initiation"], ["fvt", "video/vnd.fvt"], ["fxp", "application/vnd.adobe.fxp"], ["fxpl", "application/vnd.adobe.fxp"], ["fzs", "application/vnd.fuzzysheet"], ["g2w", "application/vnd.geoplan"], ["g3", "image/g3fax"], ["g3w", "application/vnd.geospace"], ["gac", "application/vnd.groove-account"], ["gam", "application/x-tads"], ["gbr", "application/rpki-ghostbusters"], ["gca", "application/x-gca-compressed"], ["gdl", "model/vnd.gdl"], ["gdoc", "application/vnd.google-apps.document"], ["geo", "application/vnd.dynageo"], ["geojson", "application/geo+json"], ["gex", "application/vnd.geometry-explorer"], ["ggb", "application/vnd.geogebra.file"], ["ggt", "application/vnd.geogebra.tool"], ["ghf", "application/vnd.groove-help"], ["gif", "image/gif"], ["gim", "application/vnd.groove-identity-message"], ["glb", "model/gltf-binary"], ["gltf", "model/gltf+json"], ["gml", "application/gml+xml"], ["gmx", "application/vnd.gmx"], ["gnumeric", "application/x-gnumeric"], ["gpg", "application/gpg-keys"], ["gph", "application/vnd.flographit"], ["gpx", "application/gpx+xml"], ["gqf", "application/vnd.grafeq"], ["gqs", "application/vnd.grafeq"], ["gram", "application/srgs"], ["gramps", "application/x-gramps-xml"], ["gre", "application/vnd.geometry-explorer"], ["grv", "application/vnd.groove-injector"], ["grxml", "application/srgs+xml"], ["gsf", "application/x-font-ghostscript"], ["gsheet", "application/vnd.google-apps.spreadsheet"], ["gslides", "application/vnd.google-apps.presentation"], ["gtar", "application/x-gtar"], ["gtm", "application/vnd.groove-tool-message"], ["gtw", "model/vnd.gtw"], ["gv", "text/vnd.graphviz"], ["gxf", "application/gxf"], ["gxt", "application/vnd.geonext"], ["gz", "application/gzip"], ["gzip", "application/gzip"], ["h", "text/x-c"], ["h261", "video/h261"], ["h263", "video/h263"], ["h264", "video/h264"], ["hal", "application/vnd.hal+xml"], ["hbci", "application/vnd.hbci"], ["hbs", "text/x-handlebars-template"], ["hdd", "application/x-virtualbox-hdd"], ["hdf", "application/x-hdf"], ["heic", "image/heic"], ["heics", "image/heic-sequence"], ["heif", "image/heif"], ["heifs", "image/heif-sequence"], ["hej2", "image/hej2k"], ["held", "application/atsc-held+xml"], ["hh", "text/x-c"], ["hjson", "application/hjson"], ["hlp", "application/winhlp"], ["hpgl", "application/vnd.hp-hpgl"], ["hpid", "application/vnd.hp-hpid"], ["hps", "application/vnd.hp-hps"], ["hqx", "application/mac-binhex40"], ["hsj2", "image/hsj2"], ["htc", "text/x-component"], ["htke", "application/vnd.kenameaapp"], ["htm", "text/html"], ["html", "text/html"], ["hvd", "application/vnd.yamaha.hv-dic"], ["hvp", "application/vnd.yamaha.hv-voice"], ["hvs", "application/vnd.yamaha.hv-script"], ["i2g", "application/vnd.intergeo"], ["icc", "application/vnd.iccprofile"], ["ice", "x-conference/x-cooltalk"], ["icm", "application/vnd.iccprofile"], ["ico", "image/x-icon"], ["ics", "text/calendar"], ["ief", "image/ief"], ["ifb", "text/calendar"], ["ifm", "application/vnd.shana.informed.formdata"], ["iges", "model/iges"], ["igl", "application/vnd.igloader"], ["igm", "application/vnd.insors.igm"], ["igs", "model/iges"], ["igx", "application/vnd.micrografx.igx"], ["iif", "application/vnd.shana.informed.interchange"], ["img", "application/octet-stream"], ["imp", "application/vnd.accpac.simply.imp"], ["ims", "application/vnd.ms-ims"], ["in", "text/plain"], ["ini", "text/plain"], ["ink", "application/inkml+xml"], ["inkml", "application/inkml+xml"], ["install", "application/x-install-instructions"], ["iota", "application/vnd.astraea-software.iota"], ["ipfix", "application/ipfix"], ["ipk", "application/vnd.shana.informed.package"], ["irm", "application/vnd.ibm.rights-management"], ["irp", "application/vnd.irepository.package+xml"], ["iso", "application/x-iso9660-image"], ["itp", "application/vnd.shana.informed.formtemplate"], ["its", "application/its+xml"], ["ivp", "application/vnd.immervision-ivp"], ["ivu", "application/vnd.immervision-ivu"], ["jad", "text/vnd.sun.j2me.app-descriptor"], ["jade", "text/jade"], ["jam", "application/vnd.jam"], ["jar", "application/java-archive"], ["jardiff", "application/x-java-archive-diff"], ["java", "text/x-java-source"], ["jhc", "image/jphc"], ["jisp", "application/vnd.jisp"], ["jls", "image/jls"], ["jlt", "application/vnd.hp-jlyt"], ["jng", "image/x-jng"], ["jnlp", "application/x-java-jnlp-file"], ["joda", "application/vnd.joost.joda-archive"], ["jp2", "image/jp2"], ["jpe", "image/jpeg"], ["jpeg", "image/jpeg"], ["jpf", "image/jpx"], ["jpg", "image/jpeg"], ["jpg2", "image/jp2"], ["jpgm", "video/jpm"], ["jpgv", "video/jpeg"], ["jph", "image/jph"], ["jpm", "video/jpm"], ["jpx", "image/jpx"], ["js", "application/javascript"], ["json", "application/json"], ["json5", "application/json5"], ["jsonld", "application/ld+json"], ["jsonl", "application/jsonl"], ["jsonml", "application/jsonml+json"], ["jsx", "text/jsx"], ["jxr", "image/jxr"], ["jxra", "image/jxra"], ["jxrs", "image/jxrs"], ["jxs", "image/jxs"], ["jxsc", "image/jxsc"], ["jxsi", "image/jxsi"], ["jxss", "image/jxss"], ["kar", "audio/midi"], ["karbon", "application/vnd.kde.karbon"], ["kdb", "application/octet-stream"], ["kdbx", "application/x-keepass2"], ["key", "application/x-iwork-keynote-sffkey"], ["kfo", "application/vnd.kde.kformula"], ["kia", "application/vnd.kidspiration"], ["kml", "application/vnd.google-earth.kml+xml"], ["kmz", "application/vnd.google-earth.kmz"], ["kne", "application/vnd.kinar"], ["knp", "application/vnd.kinar"], ["kon", "application/vnd.kde.kontour"], ["kpr", "application/vnd.kde.kpresenter"], ["kpt", "application/vnd.kde.kpresenter"], ["kpxx", "application/vnd.ds-keypoint"], ["ksp", "application/vnd.kde.kspread"], ["ktr", "application/vnd.kahootz"], ["ktx", "image/ktx"], ["ktx2", "image/ktx2"], ["ktz", "application/vnd.kahootz"], ["kwd", "application/vnd.kde.kword"], ["kwt", "application/vnd.kde.kword"], ["lasxml", "application/vnd.las.las+xml"], ["latex", "application/x-latex"], ["lbd", "application/vnd.llamagraphics.life-balance.desktop"], ["lbe", "application/vnd.llamagraphics.life-balance.exchange+xml"], ["les", "application/vnd.hhe.lesson-player"], ["less", "text/less"], ["lgr", "application/lgr+xml"], ["lha", "application/octet-stream"], ["link66", "application/vnd.route66.link66+xml"], ["list", "text/plain"], ["list3820", "application/vnd.ibm.modcap"], ["listafp", "application/vnd.ibm.modcap"], ["litcoffee", "text/coffeescript"], ["lnk", "application/x-ms-shortcut"], ["log", "text/plain"], ["lostxml", "application/lost+xml"], ["lrf", "application/octet-stream"], ["lrm", "application/vnd.ms-lrm"], ["ltf", "application/vnd.frogans.ltf"], ["lua", "text/x-lua"], ["luac", "application/x-lua-bytecode"], ["lvp", "audio/vnd.lucent.voice"], ["lwp", "application/vnd.lotus-wordpro"], ["lzh", "application/octet-stream"], ["m1v", "video/mpeg"], ["m2a", "audio/mpeg"], ["m2v", "video/mpeg"], ["m3a", "audio/mpeg"], ["m3u", "text/plain"], ["m3u8", "application/vnd.apple.mpegurl"], ["m4a", "audio/x-m4a"], ["m4p", "application/mp4"], ["m4s", "video/iso.segment"], ["m4u", "application/vnd.mpegurl"], ["m4v", "video/x-m4v"], ["m13", "application/x-msmediaview"], ["m14", "application/x-msmediaview"], ["m21", "application/mp21"], ["ma", "application/mathematica"], ["mads", "application/mads+xml"], ["maei", "application/mmt-aei+xml"], ["mag", "application/vnd.ecowin.chart"], ["maker", "application/vnd.framemaker"], ["man", "text/troff"], ["manifest", "text/cache-manifest"], ["map", "application/json"], ["mar", "application/octet-stream"], ["markdown", "text/markdown"], ["mathml", "application/mathml+xml"], ["mb", "application/mathematica"], ["mbk", "application/vnd.mobius.mbk"], ["mbox", "application/mbox"], ["mc1", "application/vnd.medcalcdata"], ["mcd", "application/vnd.mcd"], ["mcurl", "text/vnd.curl.mcurl"], ["md", "text/markdown"], ["mdb", "application/x-msaccess"], ["mdi", "image/vnd.ms-modi"], ["mdx", "text/mdx"], ["me", "text/troff"], ["mesh", "model/mesh"], ["meta4", "application/metalink4+xml"], ["metalink", "application/metalink+xml"], ["mets", "application/mets+xml"], ["mfm", "application/vnd.mfmp"], ["mft", "application/rpki-manifest"], ["mgp", "application/vnd.osgeo.mapguide.package"], ["mgz", "application/vnd.proteus.magazine"], ["mid", "audio/midi"], ["midi", "audio/midi"], ["mie", "application/x-mie"], ["mif", "application/vnd.mif"], ["mime", "message/rfc822"], ["mj2", "video/mj2"], ["mjp2", "video/mj2"], ["mjs", "application/javascript"], ["mk3d", "video/x-matroska"], ["mka", "audio/x-matroska"], ["mkd", "text/x-markdown"], ["mks", "video/x-matroska"], ["mkv", "video/x-matroska"], ["mlp", "application/vnd.dolby.mlp"], ["mmd", "application/vnd.chipnuts.karaoke-mmd"], ["mmf", "application/vnd.smaf"], ["mml", "text/mathml"], ["mmr", "image/vnd.fujixerox.edmics-mmr"], ["mng", "video/x-mng"], ["mny", "application/x-msmoney"], ["mobi", "application/x-mobipocket-ebook"], ["mods", "application/mods+xml"], ["mov", "video/quicktime"], ["movie", "video/x-sgi-movie"], ["mp2", "audio/mpeg"], ["mp2a", "audio/mpeg"], ["mp3", "audio/mpeg"], ["mp4", "video/mp4"], ["mp4a", "audio/mp4"], ["mp4s", "application/mp4"], ["mp4v", "video/mp4"], ["mp21", "application/mp21"], ["mpc", "application/vnd.mophun.certificate"], ["mpd", "application/dash+xml"], ["mpe", "video/mpeg"], ["mpeg", "video/mpeg"], ["mpg", "video/mpeg"], ["mpg4", "video/mp4"], ["mpga", "audio/mpeg"], ["mpkg", "application/vnd.apple.installer+xml"], ["mpm", "application/vnd.blueice.multipass"], ["mpn", "application/vnd.mophun.application"], ["mpp", "application/vnd.ms-project"], ["mpt", "application/vnd.ms-project"], ["mpy", "application/vnd.ibm.minipay"], ["mqy", "application/vnd.mobius.mqy"], ["mrc", "application/marc"], ["mrcx", "application/marcxml+xml"], ["ms", "text/troff"], ["mscml", "application/mediaservercontrol+xml"], ["mseed", "application/vnd.fdsn.mseed"], ["mseq", "application/vnd.mseq"], ["msf", "application/vnd.epson.msf"], ["msg", "application/vnd.ms-outlook"], ["msh", "model/mesh"], ["msi", "application/x-msdownload"], ["msl", "application/vnd.mobius.msl"], ["msm", "application/octet-stream"], ["msp", "application/octet-stream"], ["msty", "application/vnd.muvee.style"], ["mtl", "model/mtl"], ["mts", "model/vnd.mts"], ["mus", "application/vnd.musician"], ["musd", "application/mmt-usd+xml"], ["musicxml", "application/vnd.recordare.musicxml+xml"], ["mvb", "application/x-msmediaview"], ["mvt", "application/vnd.mapbox-vector-tile"], ["mwf", "application/vnd.mfer"], ["mxf", "application/mxf"], ["mxl", "application/vnd.recordare.musicxml"], ["mxmf", "audio/mobile-xmf"], ["mxml", "application/xv+xml"], ["mxs", "application/vnd.triscape.mxs"], ["mxu", "video/vnd.mpegurl"], ["n-gage", "application/vnd.nokia.n-gage.symbian.install"], ["n3", "text/n3"], ["nb", "application/mathematica"], ["nbp", "application/vnd.wolfram.player"], ["nc", "application/x-netcdf"], ["ncx", "application/x-dtbncx+xml"], ["nfo", "text/x-nfo"], ["ngdat", "application/vnd.nokia.n-gage.data"], ["nitf", "application/vnd.nitf"], ["nlu", "application/vnd.neurolanguage.nlu"], ["nml", "application/vnd.enliven"], ["nnd", "application/vnd.noblenet-directory"], ["nns", "application/vnd.noblenet-sealer"], ["nnw", "application/vnd.noblenet-web"], ["npx", "image/vnd.net-fpx"], ["nq", "application/n-quads"], ["nsc", "application/x-conference"], ["nsf", "application/vnd.lotus-notes"], ["nt", "application/n-triples"], ["ntf", "application/vnd.nitf"], ["numbers", "application/x-iwork-numbers-sffnumbers"], ["nzb", "application/x-nzb"], ["oa2", "application/vnd.fujitsu.oasys2"], ["oa3", "application/vnd.fujitsu.oasys3"], ["oas", "application/vnd.fujitsu.oasys"], ["obd", "application/x-msbinder"], ["obgx", "application/vnd.openblox.game+xml"], ["obj", "model/obj"], ["oda", "application/oda"], ["odb", "application/vnd.oasis.opendocument.database"], ["odc", "application/vnd.oasis.opendocument.chart"], ["odf", "application/vnd.oasis.opendocument.formula"], ["odft", "application/vnd.oasis.opendocument.formula-template"], ["odg", "application/vnd.oasis.opendocument.graphics"], ["odi", "application/vnd.oasis.opendocument.image"], ["odm", "application/vnd.oasis.opendocument.text-master"], ["odp", "application/vnd.oasis.opendocument.presentation"], ["ods", "application/vnd.oasis.opendocument.spreadsheet"], ["odt", "application/vnd.oasis.opendocument.text"], ["oga", "audio/ogg"], ["ogex", "model/vnd.opengex"], ["ogg", "audio/ogg"], ["ogv", "video/ogg"], ["ogx", "application/ogg"], ["omdoc", "application/omdoc+xml"], ["onepkg", "application/onenote"], ["onetmp", "application/onenote"], ["onetoc", "application/onenote"], ["onetoc2", "application/onenote"], ["opf", "application/oebps-package+xml"], ["opml", "text/x-opml"], ["oprc", "application/vnd.palm"], ["opus", "audio/ogg"], ["org", "text/x-org"], ["osf", "application/vnd.yamaha.openscoreformat"], ["osfpvg", "application/vnd.yamaha.openscoreformat.osfpvg+xml"], ["osm", "application/vnd.openstreetmap.data+xml"], ["otc", "application/vnd.oasis.opendocument.chart-template"], ["otf", "font/otf"], ["otg", "application/vnd.oasis.opendocument.graphics-template"], ["oth", "application/vnd.oasis.opendocument.text-web"], ["oti", "application/vnd.oasis.opendocument.image-template"], ["otp", "application/vnd.oasis.opendocument.presentation-template"], ["ots", "application/vnd.oasis.opendocument.spreadsheet-template"], ["ott", "application/vnd.oasis.opendocument.text-template"], ["ova", "application/x-virtualbox-ova"], ["ovf", "application/x-virtualbox-ovf"], ["owl", "application/rdf+xml"], ["oxps", "application/oxps"], ["oxt", "application/vnd.openofficeorg.extension"], ["p", "text/x-pascal"], ["p7a", "application/x-pkcs7-signature"], ["p7b", "application/x-pkcs7-certificates"], ["p7c", "application/pkcs7-mime"], ["p7m", "application/pkcs7-mime"], ["p7r", "application/x-pkcs7-certreqresp"], ["p7s", "application/pkcs7-signature"], ["p8", "application/pkcs8"], ["p10", "application/x-pkcs10"], ["p12", "application/x-pkcs12"], ["pac", "application/x-ns-proxy-autoconfig"], ["pages", "application/x-iwork-pages-sffpages"], ["pas", "text/x-pascal"], ["paw", "application/vnd.pawaafile"], ["pbd", "application/vnd.powerbuilder6"], ["pbm", "image/x-portable-bitmap"], ["pcap", "application/vnd.tcpdump.pcap"], ["pcf", "application/x-font-pcf"], ["pcl", "application/vnd.hp-pcl"], ["pclxl", "application/vnd.hp-pclxl"], ["pct", "image/x-pict"], ["pcurl", "application/vnd.curl.pcurl"], ["pcx", "image/x-pcx"], ["pdb", "application/x-pilot"], ["pde", "text/x-processing"], ["pdf", "application/pdf"], ["pem", "application/x-x509-user-cert"], ["pfa", "application/x-font-type1"], ["pfb", "application/x-font-type1"], ["pfm", "application/x-font-type1"], ["pfr", "application/font-tdpfr"], ["pfx", "application/x-pkcs12"], ["pgm", "image/x-portable-graymap"], ["pgn", "application/x-chess-pgn"], ["pgp", "application/pgp"], ["php", "application/x-httpd-php"], ["php3", "application/x-httpd-php"], ["php4", "application/x-httpd-php"], ["phps", "application/x-httpd-php-source"], ["phtml", "application/x-httpd-php"], ["pic", "image/x-pict"], ["pkg", "application/octet-stream"], ["pki", "application/pkixcmp"], ["pkipath", "application/pkix-pkipath"], ["pkpass", "application/vnd.apple.pkpass"], ["pl", "application/x-perl"], ["plb", "application/vnd.3gpp.pic-bw-large"], ["plc", "application/vnd.mobius.plc"], ["plf", "application/vnd.pocketlearn"], ["pls", "application/pls+xml"], ["pm", "application/x-perl"], ["pml", "application/vnd.ctc-posml"], ["png", "image/png"], ["pnm", "image/x-portable-anymap"], ["portpkg", "application/vnd.macports.portpkg"], ["pot", "application/vnd.ms-powerpoint"], ["potm", "application/vnd.ms-powerpoint.presentation.macroEnabled.12"], ["potx", "application/vnd.openxmlformats-officedocument.presentationml.template"], ["ppa", "application/vnd.ms-powerpoint"], ["ppam", "application/vnd.ms-powerpoint.addin.macroEnabled.12"], ["ppd", "application/vnd.cups-ppd"], ["ppm", "image/x-portable-pixmap"], ["pps", "application/vnd.ms-powerpoint"], ["ppsm", "application/vnd.ms-powerpoint.slideshow.macroEnabled.12"], ["ppsx", "application/vnd.openxmlformats-officedocument.presentationml.slideshow"], ["ppt", "application/powerpoint"], ["pptm", "application/vnd.ms-powerpoint.presentation.macroEnabled.12"], ["pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"], ["pqa", "application/vnd.palm"], ["prc", "application/x-pilot"], ["pre", "application/vnd.lotus-freelance"], ["prf", "application/pics-rules"], ["provx", "application/provenance+xml"], ["ps", "application/postscript"], ["psb", "application/vnd.3gpp.pic-bw-small"], ["psd", "application/x-photoshop"], ["psf", "application/x-font-linux-psf"], ["pskcxml", "application/pskc+xml"], ["pti", "image/prs.pti"], ["ptid", "application/vnd.pvi.ptid1"], ["pub", "application/x-mspublisher"], ["pvb", "application/vnd.3gpp.pic-bw-var"], ["pwn", "application/vnd.3m.post-it-notes"], ["pya", "audio/vnd.ms-playready.media.pya"], ["pyv", "video/vnd.ms-playready.media.pyv"], ["qam", "application/vnd.epson.quickanime"], ["qbo", "application/vnd.intu.qbo"], ["qfx", "application/vnd.intu.qfx"], ["qps", "application/vnd.publishare-delta-tree"], ["qt", "video/quicktime"], ["qwd", "application/vnd.quark.quarkxpress"], ["qwt", "application/vnd.quark.quarkxpress"], ["qxb", "application/vnd.quark.quarkxpress"], ["qxd", "application/vnd.quark.quarkxpress"], ["qxl", "application/vnd.quark.quarkxpress"], ["qxt", "application/vnd.quark.quarkxpress"], ["ra", "audio/x-realaudio"], ["ram", "audio/x-pn-realaudio"], ["raml", "application/raml+yaml"], ["rapd", "application/route-apd+xml"], ["rar", "application/x-rar"], ["ras", "image/x-cmu-raster"], ["rcprofile", "application/vnd.ipunplugged.rcprofile"], ["rdf", "application/rdf+xml"], ["rdz", "application/vnd.data-vision.rdz"], ["relo", "application/p2p-overlay+xml"], ["rep", "application/vnd.businessobjects"], ["res", "application/x-dtbresource+xml"], ["rgb", "image/x-rgb"], ["rif", "application/reginfo+xml"], ["rip", "audio/vnd.rip"], ["ris", "application/x-research-info-systems"], ["rl", "application/resource-lists+xml"], ["rlc", "image/vnd.fujixerox.edmics-rlc"], ["rld", "application/resource-lists-diff+xml"], ["rm", "audio/x-pn-realaudio"], ["rmi", "audio/midi"], ["rmp", "audio/x-pn-realaudio-plugin"], ["rms", "application/vnd.jcp.javame.midlet-rms"], ["rmvb", "application/vnd.rn-realmedia-vbr"], ["rnc", "application/relax-ng-compact-syntax"], ["rng", "application/xml"], ["roa", "application/rpki-roa"], ["roff", "text/troff"], ["rp9", "application/vnd.cloanto.rp9"], ["rpm", "audio/x-pn-realaudio-plugin"], ["rpss", "application/vnd.nokia.radio-presets"], ["rpst", "application/vnd.nokia.radio-preset"], ["rq", "application/sparql-query"], ["rs", "application/rls-services+xml"], ["rsa", "application/x-pkcs7"], ["rsat", "application/atsc-rsat+xml"], ["rsd", "application/rsd+xml"], ["rsheet", "application/urc-ressheet+xml"], ["rss", "application/rss+xml"], ["rtf", "text/rtf"], ["rtx", "text/richtext"], ["run", "application/x-makeself"], ["rusd", "application/route-usd+xml"], ["rv", "video/vnd.rn-realvideo"], ["s", "text/x-asm"], ["s3m", "audio/s3m"], ["saf", "application/vnd.yamaha.smaf-audio"], ["sass", "text/x-sass"], ["sbml", "application/sbml+xml"], ["sc", "application/vnd.ibm.secure-container"], ["scd", "application/x-msschedule"], ["scm", "application/vnd.lotus-screencam"], ["scq", "application/scvp-cv-request"], ["scs", "application/scvp-cv-response"], ["scss", "text/x-scss"], ["scurl", "text/vnd.curl.scurl"], ["sda", "application/vnd.stardivision.draw"], ["sdc", "application/vnd.stardivision.calc"], ["sdd", "application/vnd.stardivision.impress"], ["sdkd", "application/vnd.solent.sdkm+xml"], ["sdkm", "application/vnd.solent.sdkm+xml"], ["sdp", "application/sdp"], ["sdw", "application/vnd.stardivision.writer"], ["sea", "application/octet-stream"], ["see", "application/vnd.seemail"], ["seed", "application/vnd.fdsn.seed"], ["sema", "application/vnd.sema"], ["semd", "application/vnd.semd"], ["semf", "application/vnd.semf"], ["senmlx", "application/senml+xml"], ["sensmlx", "application/sensml+xml"], ["ser", "application/java-serialized-object"], ["setpay", "application/set-payment-initiation"], ["setreg", "application/set-registration-initiation"], ["sfd-hdstx", "application/vnd.hydrostatix.sof-data"], ["sfs", "application/vnd.spotfire.sfs"], ["sfv", "text/x-sfv"], ["sgi", "image/sgi"], ["sgl", "application/vnd.stardivision.writer-global"], ["sgm", "text/sgml"], ["sgml", "text/sgml"], ["sh", "application/x-sh"], ["shar", "application/x-shar"], ["shex", "text/shex"], ["shf", "application/shf+xml"], ["shtml", "text/html"], ["sid", "image/x-mrsid-image"], ["sieve", "application/sieve"], ["sig", "application/pgp-signature"], ["sil", "audio/silk"], ["silo", "model/mesh"], ["sis", "application/vnd.symbian.install"], ["sisx", "application/vnd.symbian.install"], ["sit", "application/x-stuffit"], ["sitx", "application/x-stuffitx"], ["siv", "application/sieve"], ["skd", "application/vnd.koan"], ["skm", "application/vnd.koan"], ["skp", "application/vnd.koan"], ["skt", "application/vnd.koan"], ["sldm", "application/vnd.ms-powerpoint.slide.macroenabled.12"], ["sldx", "application/vnd.openxmlformats-officedocument.presentationml.slide"], ["slim", "text/slim"], ["slm", "text/slim"], ["sls", "application/route-s-tsid+xml"], ["slt", "application/vnd.epson.salt"], ["sm", "application/vnd.stepmania.stepchart"], ["smf", "application/vnd.stardivision.math"], ["smi", "application/smil"], ["smil", "application/smil"], ["smv", "video/x-smv"], ["smzip", "application/vnd.stepmania.package"], ["snd", "audio/basic"], ["snf", "application/x-font-snf"], ["so", "application/octet-stream"], ["spc", "application/x-pkcs7-certificates"], ["spdx", "text/spdx"], ["spf", "application/vnd.yamaha.smaf-phrase"], ["spl", "application/x-futuresplash"], ["spot", "text/vnd.in3d.spot"], ["spp", "application/scvp-vp-response"], ["spq", "application/scvp-vp-request"], ["spx", "audio/ogg"], ["sql", "application/x-sql"], ["src", "application/x-wais-source"], ["srt", "application/x-subrip"], ["sru", "application/sru+xml"], ["srx", "application/sparql-results+xml"], ["ssdl", "application/ssdl+xml"], ["sse", "application/vnd.kodak-descriptor"], ["ssf", "application/vnd.epson.ssf"], ["ssml", "application/ssml+xml"], ["sst", "application/octet-stream"], ["st", "application/vnd.sailingtracker.track"], ["stc", "application/vnd.sun.xml.calc.template"], ["std", "application/vnd.sun.xml.draw.template"], ["stf", "application/vnd.wt.stf"], ["sti", "application/vnd.sun.xml.impress.template"], ["stk", "application/hyperstudio"], ["stl", "model/stl"], ["stpx", "model/step+xml"], ["stpxz", "model/step-xml+zip"], ["stpz", "model/step+zip"], ["str", "application/vnd.pg.format"], ["stw", "application/vnd.sun.xml.writer.template"], ["styl", "text/stylus"], ["stylus", "text/stylus"], ["sub", "text/vnd.dvb.subtitle"], ["sus", "application/vnd.sus-calendar"], ["susp", "application/vnd.sus-calendar"], ["sv4cpio", "application/x-sv4cpio"], ["sv4crc", "application/x-sv4crc"], ["svc", "application/vnd.dvb.service"], ["svd", "application/vnd.svd"], ["svg", "image/svg+xml"], ["svgz", "image/svg+xml"], ["swa", "application/x-director"], ["swf", "application/x-shockwave-flash"], ["swi", "application/vnd.aristanetworks.swi"], ["swidtag", "application/swid+xml"], ["sxc", "application/vnd.sun.xml.calc"], ["sxd", "application/vnd.sun.xml.draw"], ["sxg", "application/vnd.sun.xml.writer.global"], ["sxi", "application/vnd.sun.xml.impress"], ["sxm", "application/vnd.sun.xml.math"], ["sxw", "application/vnd.sun.xml.writer"], ["t", "text/troff"], ["t3", "application/x-t3vm-image"], ["t38", "image/t38"], ["taglet", "application/vnd.mynfc"], ["tao", "application/vnd.tao.intent-module-archive"], ["tap", "image/vnd.tencent.tap"], ["tar", "application/x-tar"], ["tcap", "application/vnd.3gpp2.tcap"], ["tcl", "application/x-tcl"], ["td", "application/urc-targetdesc+xml"], ["teacher", "application/vnd.smart.teacher"], ["tei", "application/tei+xml"], ["teicorpus", "application/tei+xml"], ["tex", "application/x-tex"], ["texi", "application/x-texinfo"], ["texinfo", "application/x-texinfo"], ["text", "text/plain"], ["tfi", "application/thraud+xml"], ["tfm", "application/x-tex-tfm"], ["tfx", "image/tiff-fx"], ["tga", "image/x-tga"], ["tgz", "application/x-tar"], ["thmx", "application/vnd.ms-officetheme"], ["tif", "image/tiff"], ["tiff", "image/tiff"], ["tk", "application/x-tcl"], ["tmo", "application/vnd.tmobile-livetv"], ["toml", "application/toml"], ["torrent", "application/x-bittorrent"], ["tpl", "application/vnd.groove-tool-template"], ["tpt", "application/vnd.trid.tpt"], ["tr", "text/troff"], ["tra", "application/vnd.trueapp"], ["trig", "application/trig"], ["trm", "application/x-msterminal"], ["ts", "video/mp2t"], ["tsd", "application/timestamped-data"], ["tsv", "text/tab-separated-values"], ["ttc", "font/collection"], ["ttf", "font/ttf"], ["ttl", "text/turtle"], ["ttml", "application/ttml+xml"], ["twd", "application/vnd.simtech-mindmapper"], ["twds", "application/vnd.simtech-mindmapper"], ["txd", "application/vnd.genomatix.tuxedo"], ["txf", "application/vnd.mobius.txf"], ["txt", "text/plain"], ["u8dsn", "message/global-delivery-status"], ["u8hdr", "message/global-headers"], ["u8mdn", "message/global-disposition-notification"], ["u8msg", "message/global"], ["u32", "application/x-authorware-bin"], ["ubj", "application/ubjson"], ["udeb", "application/x-debian-package"], ["ufd", "application/vnd.ufdl"], ["ufdl", "application/vnd.ufdl"], ["ulx", "application/x-glulx"], ["umj", "application/vnd.umajin"], ["unityweb", "application/vnd.unity"], ["uoml", "application/vnd.uoml+xml"], ["uri", "text/uri-list"], ["uris", "text/uri-list"], ["urls", "text/uri-list"], ["usdz", "model/vnd.usdz+zip"], ["ustar", "application/x-ustar"], ["utz", "application/vnd.uiq.theme"], ["uu", "text/x-uuencode"], ["uva", "audio/vnd.dece.audio"], ["uvd", "application/vnd.dece.data"], ["uvf", "application/vnd.dece.data"], ["uvg", "image/vnd.dece.graphic"], ["uvh", "video/vnd.dece.hd"], ["uvi", "image/vnd.dece.graphic"], ["uvm", "video/vnd.dece.mobile"], ["uvp", "video/vnd.dece.pd"], ["uvs", "video/vnd.dece.sd"], ["uvt", "application/vnd.dece.ttml+xml"], ["uvu", "video/vnd.uvvu.mp4"], ["uvv", "video/vnd.dece.video"], ["uvva", "audio/vnd.dece.audio"], ["uvvd", "application/vnd.dece.data"], ["uvvf", "application/vnd.dece.data"], ["uvvg", "image/vnd.dece.graphic"], ["uvvh", "video/vnd.dece.hd"], ["uvvi", "image/vnd.dece.graphic"], ["uvvm", "video/vnd.dece.mobile"], ["uvvp", "video/vnd.dece.pd"], ["uvvs", "video/vnd.dece.sd"], ["uvvt", "application/vnd.dece.ttml+xml"], ["uvvu", "video/vnd.uvvu.mp4"], ["uvvv", "video/vnd.dece.video"], ["uvvx", "application/vnd.dece.unspecified"], ["uvvz", "application/vnd.dece.zip"], ["uvx", "application/vnd.dece.unspecified"], ["uvz", "application/vnd.dece.zip"], ["vbox", "application/x-virtualbox-vbox"], ["vbox-extpack", "application/x-virtualbox-vbox-extpack"], ["vcard", "text/vcard"], ["vcd", "application/x-cdlink"], ["vcf", "text/x-vcard"], ["vcg", "application/vnd.groove-vcard"], ["vcs", "text/x-vcalendar"], ["vcx", "application/vnd.vcx"], ["vdi", "application/x-virtualbox-vdi"], ["vds", "model/vnd.sap.vds"], ["vhd", "application/x-virtualbox-vhd"], ["vis", "application/vnd.visionary"], ["viv", "video/vnd.vivo"], ["vlc", "application/videolan"], ["vmdk", "application/x-virtualbox-vmdk"], ["vob", "video/x-ms-vob"], ["vor", "application/vnd.stardivision.writer"], ["vox", "application/x-authorware-bin"], ["vrml", "model/vrml"], ["vsd", "application/vnd.visio"], ["vsf", "application/vnd.vsf"], ["vss", "application/vnd.visio"], ["vst", "application/vnd.visio"], ["vsw", "application/vnd.visio"], ["vtf", "image/vnd.valve.source.texture"], ["vtt", "text/vtt"], ["vtu", "model/vnd.vtu"], ["vxml", "application/voicexml+xml"], ["w3d", "application/x-director"], ["wad", "application/x-doom"], ["wadl", "application/vnd.sun.wadl+xml"], ["war", "application/java-archive"], ["wasm", "application/wasm"], ["wav", "audio/x-wav"], ["wax", "audio/x-ms-wax"], ["wbmp", "image/vnd.wap.wbmp"], ["wbs", "application/vnd.criticaltools.wbs+xml"], ["wbxml", "application/wbxml"], ["wcm", "application/vnd.ms-works"], ["wdb", "application/vnd.ms-works"], ["wdp", "image/vnd.ms-photo"], ["weba", "audio/webm"], ["webapp", "application/x-web-app-manifest+json"], ["webm", "video/webm"], ["webmanifest", "application/manifest+json"], ["webp", "image/webp"], ["wg", "application/vnd.pmi.widget"], ["wgt", "application/widget"], ["wks", "application/vnd.ms-works"], ["wm", "video/x-ms-wm"], ["wma", "audio/x-ms-wma"], ["wmd", "application/x-ms-wmd"], ["wmf", "image/wmf"], ["wml", "text/vnd.wap.wml"], ["wmlc", "application/wmlc"], ["wmls", "text/vnd.wap.wmlscript"], ["wmlsc", "application/vnd.wap.wmlscriptc"], ["wmv", "video/x-ms-wmv"], ["wmx", "video/x-ms-wmx"], ["wmz", "application/x-msmetafile"], ["woff", "font/woff"], ["woff2", "font/woff2"], ["word", "application/msword"], ["wpd", "application/vnd.wordperfect"], ["wpl", "application/vnd.ms-wpl"], ["wps", "application/vnd.ms-works"], ["wqd", "application/vnd.wqd"], ["wri", "application/x-mswrite"], ["wrl", "model/vrml"], ["wsc", "message/vnd.wfa.wsc"], ["wsdl", "application/wsdl+xml"], ["wspolicy", "application/wspolicy+xml"], ["wtb", "application/vnd.webturbo"], ["wvx", "video/x-ms-wvx"], ["x3d", "model/x3d+xml"], ["x3db", "model/x3d+fastinfoset"], ["x3dbz", "model/x3d+binary"], ["x3dv", "model/x3d-vrml"], ["x3dvz", "model/x3d+vrml"], ["x3dz", "model/x3d+xml"], ["x32", "application/x-authorware-bin"], ["x_b", "model/vnd.parasolid.transmit.binary"], ["x_t", "model/vnd.parasolid.transmit.text"], ["xaml", "application/xaml+xml"], ["xap", "application/x-silverlight-app"], ["xar", "application/vnd.xara"], ["xav", "application/xcap-att+xml"], ["xbap", "application/x-ms-xbap"], ["xbd", "application/vnd.fujixerox.docuworks.binder"], ["xbm", "image/x-xbitmap"], ["xca", "application/xcap-caps+xml"], ["xcs", "application/calendar+xml"], ["xdf", "application/xcap-diff+xml"], ["xdm", "application/vnd.syncml.dm+xml"], ["xdp", "application/vnd.adobe.xdp+xml"], ["xdssc", "application/dssc+xml"], ["xdw", "application/vnd.fujixerox.docuworks"], ["xel", "application/xcap-el+xml"], ["xenc", "application/xenc+xml"], ["xer", "application/patch-ops-error+xml"], ["xfdf", "application/vnd.adobe.xfdf"], ["xfdl", "application/vnd.xfdl"], ["xht", "application/xhtml+xml"], ["xhtml", "application/xhtml+xml"], ["xhvml", "application/xv+xml"], ["xif", "image/vnd.xiff"], ["xl", "application/excel"], ["xla", "application/vnd.ms-excel"], ["xlam", "application/vnd.ms-excel.addin.macroEnabled.12"], ["xlc", "application/vnd.ms-excel"], ["xlf", "application/xliff+xml"], ["xlm", "application/vnd.ms-excel"], ["xls", "application/vnd.ms-excel"], ["xlsb", "application/vnd.ms-excel.sheet.binary.macroEnabled.12"], ["xlsm", "application/vnd.ms-excel.sheet.macroEnabled.12"], ["xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"], ["xlt", "application/vnd.ms-excel"], ["xltm", "application/vnd.ms-excel.template.macroEnabled.12"], ["xltx", "application/vnd.openxmlformats-officedocument.spreadsheetml.template"], ["xlw", "application/vnd.ms-excel"], ["xm", "audio/xm"], ["xml", "application/xml"], ["xns", "application/xcap-ns+xml"], ["xo", "application/vnd.olpc-sugar"], ["xop", "application/xop+xml"], ["xpi", "application/x-xpinstall"], ["xpl", "application/xproc+xml"], ["xpm", "image/x-xpixmap"], ["xpr", "application/vnd.is-xpr"], ["xps", "application/vnd.ms-xpsdocument"], ["xpw", "application/vnd.intercon.formnet"], ["xpx", "application/vnd.intercon.formnet"], ["xsd", "application/xml"], ["xsl", "application/xml"], ["xslt", "application/xslt+xml"], ["xsm", "application/vnd.syncml+xml"], ["xspf", "application/xspf+xml"], ["xul", "application/vnd.mozilla.xul+xml"], ["xvm", "application/xv+xml"], ["xvml", "application/xv+xml"], ["xwd", "image/x-xwindowdump"], ["xyz", "chemical/x-xyz"], ["xz", "application/x-xz"], ["yaml", "text/yaml"], ["yang", "application/yang"], ["yin", "application/yin+xml"], ["yml", "text/yaml"], ["ymp", "text/x-suse-ymp"], ["z", "application/x-compress"], ["z1", "application/x-zmachine"], ["z2", "application/x-zmachine"], ["z3", "application/x-zmachine"], ["z4", "application/x-zmachine"], ["z5", "application/x-zmachine"], ["z6", "application/x-zmachine"], ["z7", "application/x-zmachine"], ["z8", "application/x-zmachine"], ["zaz", "application/vnd.zzazz.deck+xml"], ["zip", "application/zip"], ["zir", "application/vnd.zul"], ["zirz", "application/vnd.zul"], ["zmm", "application/vnd.handheld-entertainment+xml"], ["zsh", "text/x-scriptzsh"]]);
function fa(e, t, a) {
    const n = sc(e)
      , {webkitRelativePath: o} = e
      , i = typeof t == "string" ? t : typeof o == "string" && o.length > 0 ? o : `./${e.name}`;
    return typeof n.path != "string" && ko(n, "path", i),
    ko(n, "relativePath", i),
    n;
}
function sc(e) {
    const {name: t} = e;
    if (t && t.lastIndexOf(".") !== -1 && !e.type) {
        const n = t.split(".").pop().toLowerCase()
          , o = lc.get(n);
        o && Object.defineProperty(e, "type", {
            value: o,
            writable: false,
            configurable: false,
            enumerable: true
        });
    }
    return e;
}
function ko(e, t, a) {
    Object.defineProperty(e, t, {
        value: a,
        writable: false,
        configurable: false,
        enumerable: true
    });
}
const cc = [".DS_Store", "Thumbs.db"];
function pc(e) {
    return na(this, void 0, void 0, function*() {
        return Ja(e) && dc(e.dataTransfer) ? vc(e.dataTransfer, e.type) : uc(e) ? mc(e) : Array.isArray(e) && e.every( (t) => "getFile"in t && typeof t.getFile == "function") ? fc(e) : [];
    });
}
function dc(e) {
    return Ja(e);
}
function uc(e) {
    return Ja(e) && Ja(e.target);
}
function Ja(e) {
    return typeof e == "object" && e !== null;
}
function mc(e) {
    return En(e.target.files).map( (t) => fa(t));
}
function fc(e) {
    return na(this, void 0, void 0, function*() {
        return (yield Promise.all(e.map( (a) => a.getFile()))).map( (a) => fa(a));
    });
}
function vc(e, t) {
    return na(this, void 0, void 0, function*() {
        if (e.items) {
            const a = En(e.items).filter( (o) => o.kind === "file");
            if (t !== "drop")
                return a;
            const n = yield Promise.all(a.map(gc));
            return To(bi(n));
        }
        return To(En(e.files).map( (a) => fa(a)));
    });
}
function To(e) {
    return e.filter( (t) => cc.indexOf(t.name) === -1);
}
function En(e) {
    if (e === null)
        return [];
    const t = [];
    for (let a = 0; a < e.length; a++) {
        const n = e[a];
        t.push(n);
    }
    return t;
}
function gc(e) {
    if (typeof e.webkitGetAsEntry != "function")
        return Po(e);
    const t = e.webkitGetAsEntry();
    return t && t.isDirectory ? yi(t) : Po(e, t);
}
function bi(e) {
    return e.reduce( (t, a) => [...t, ...Array.isArray(a) ? bi(a) : [a]], []);
}
function Po(e, t) {
    return na(this, void 0, void 0, function*() {
        var a;
        if (globalThis.isSecureContext && typeof e.getAsFileSystemHandle == "function") {
            const i = yield e.getAsFileSystemHandle();
            if (i === null)
                throw new Error(`${e} is not a File`);
            if (i !== void 0) {
                const l = yield i.getFile();
                return l.handle = i,
                fa(l);
            }
        }
        const n = e.getAsFile();
        if (!n)
            throw new Error(`${e} is not a File`);
        return fa(n, (a = t == null ? void 0 : t.fullPath) !== null && a !== void 0 ? a : void 0);
    });
}
function xc(e) {
    return na(this, void 0, void 0, function*() {
        return e.isDirectory ? yi(e) : hc(e);
    });
}
function yi(e) {
    const t = e.createReader();
    return new Promise( (a, n) => {
        const o = [];
        function i() {
            t.readEntries( (l) => na(this, void 0, void 0, function*() {
                if (l.length) {
                    const s = Promise.all(l.map(xc));
                    o.push(s),
                    i();
                } else
                    try {
                        const s = yield Promise.all(o);
                        a(s);
                    } catch (s) {
                        n(s);
                    }
            }), (l) => {
                n(l);
            }
            );
        }
        i();
    }
    );
}
function hc(e) {
    return na(this, void 0, void 0, function*() {
        return new Promise( (t, a) => {
            e.file( (n) => {
                const o = fa(n, e.fullPath);
                t(o);
            }
            , (n) => {
                a(n);
            }
            );
        }
        );
    });
}
var Ua = {}, Oo;
function bc() {
    return Oo || (Oo = 1,
    Ua.__esModule = true,
    Ua.default = function(e, t) {
        if (e && t) {
            var a = Array.isArray(t) ? t : t.split(",");
            if (a.length === 0)
                return true;
            var n = e.name || ""
              , o = (e.type || "").toLowerCase()
              , i = o.replace(/\/.*$/, "");
            return a.some(function(l) {
                var s = l.trim().toLowerCase();
                return s.charAt(0) === "." ? n.toLowerCase().endsWith(s) : s.endsWith("/*") ? i === s.replace(/\/.*$/, "") : o === s;
            });
        }
        return true;
    }
    ),
    Ua;
}
var yc = bc();
const mn = Tr(yc);
function jo(e) {
    return Sc(e) || Cc(e) || Ci(e) || wc();
}
function wc() {
    throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`);
}
function Cc(e) {
    if (typeof Symbol < "u" && e[Symbol.iterator] != null || e["@@iterator"] != null)
        return Array.from(e);
}
function Sc(e) {
    if (Array.isArray(e))
        return An(e);
}
function Io(e, t) {
    var a = Object.keys(e);
    if (Object.getOwnPropertySymbols) {
        var n = Object.getOwnPropertySymbols(e);
        t && (n = n.filter(function(o) {
            return Object.getOwnPropertyDescriptor(e, o).enumerable;
        })),
        a.push.apply(a, n);
    }
    return a;
}
function _o(e) {
    for (var t = 1; t < arguments.length; t++) {
        var a = arguments[t] != null ? arguments[t] : {};
        t % 2 ? Io(Object(a), true).forEach(function(n) {
            wi(e, n, a[n]);
        }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(a)) : Io(Object(a)).forEach(function(n) {
            Object.defineProperty(e, n, Object.getOwnPropertyDescriptor(a, n));
        });
    }
    return e;
}
function wi(e, t, a) {
    return t in e ? Object.defineProperty(e, t, {
        value: a,
        enumerable: true,
        configurable: true,
        writable: true
    }) : e[t] = a,
    e;
}
function ja(e, t) {
    return kc(e) || Ac(e, t) || Ci(e, t) || Ec();
}
function Ec() {
    throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`);
}
function Ci(e, t) {
    if (e) {
        if (typeof e == "string")
            return An(e, t);
        var a = Object.prototype.toString.call(e).slice(8, -1);
        if (a === "Object" && e.constructor && (a = e.constructor.name),
        a === "Map" || a === "Set")
            return Array.from(e);
        if (a === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(a))
            return An(e, t);
    }
}
function An(e, t) {
    (t == null || t > e.length) && (t = e.length);
    for (var a = 0, n = new Array(t); a < t; a++)
        n[a] = e[a];
    return n;
}
function Ac(e, t) {
    var a = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
    if (a != null) {
        var n = [], o = true, i = false, l, s;
        try {
            for (a = a.call(e); !(o = (l = a.next()).done) && (n.push(l.value),
            !(t && n.length === t)); o = true)
                ;
        } catch (c) {
            i = true,
            s = c;
        } finally {
            try {
                !o && a.return != null && a.return();
            } finally {
                if (i)
                    throw s;
            }
        }
        return n;
    }
}
function kc(e) {
    if (Array.isArray(e))
        return e;
}
var Tc = typeof mn == "function" ? mn : mn.default
  , Pc = "file-invalid-type"
  , Oc = "file-too-large"
  , jc = "file-too-small"
  , Ic = "too-many-files"
  , _c = function() {
    var t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : ""
      , a = t.split(",")
      , n = a.length > 1 ? "one of ".concat(a.join(", ")) : a[0];
    return {
        code: Pc,
        message: "File type must be ".concat(n)
    };
}
  , Ro = function(t) {
    return {
        code: Oc,
        message: "File is larger than ".concat(t, " ").concat(t === 1 ? "byte" : "bytes")
    };
}
  , Do = function(t) {
    return {
        code: jc,
        message: "File is smaller than ".concat(t, " ").concat(t === 1 ? "byte" : "bytes")
    };
}
  , Rc = {
    code: Ic,
    message: "Too many files"
};
function Si(e, t) {
    var a = e.type === "application/x-moz-file" || Tc(e, t);
    return [a, a ? null : _c(t)];
}
function Ei(e, t, a) {
    if (Xt(e.size))
        if (Xt(t) && Xt(a)) {
            if (e.size > a)
                return [false, Ro(a)];
            if (e.size < t)
                return [false, Do(t)];
        } else {
            if (Xt(t) && e.size < t)
                return [false, Do(t)];
            if (Xt(a) && e.size > a)
                return [false, Ro(a)];
        }
    return [true, null];
}
function Xt(e) {
    return e != null;
}
function Dc(e) {
    var t = e.files
      , a = e.accept
      , n = e.minSize
      , o = e.maxSize
      , i = e.multiple
      , l = e.maxFiles
      , s = e.validator;
    return !i && t.length > 1 || i && l >= 1 && t.length > l ? false : t.every(function(c) {
        var d = Si(c, a)
          , p = ja(d, 1)
          , u = p[0]
          , f = Ei(c, n, o)
          , m = ja(f, 1)
          , b = m[0]
          , y = s ? s(c) : null;
        return u && b && !y;
    });
}
function en(e) {
    return typeof e.isPropagationStopped == "function" ? e.isPropagationStopped() : typeof e.cancelBubble < "u" ? e.cancelBubble : false;
}
function qa(e) {
    return e.dataTransfer ? Array.prototype.some.call(e.dataTransfer.types, function(t) {
        return t === "Files" || t === "application/x-moz-file";
    }) : !!e.target && !!e.target.files;
}
function Fo(e) {
    e.preventDefault();
}
function Fc(e) {
    return e.indexOf("MSIE") !== -1 || e.indexOf("Trident/") !== -1;
}
function Lc(e) {
    return e.indexOf("Edge/") !== -1;
}
function Mc() {
    var e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : window.navigator.userAgent;
    return Fc(e) || Lc(e);
}
function Pt() {
    for (var e = arguments.length, t = new Array(e), a = 0; a < e; a++)
        t[a] = arguments[a];
    return function(n) {
        for (var o = arguments.length, i = new Array(o > 1 ? o - 1 : 0), l = 1; l < o; l++)
            i[l - 1] = arguments[l];
        return t.some(function(s) {
            return !en(n) && s && s.apply(void 0, [n].concat(i)),
            en(n);
        });
    }
    ;
}
function $c() {
    return "showOpenFilePicker"in window;
}
function Bc(e) {
    if (Xt(e)) {
        var t = Object.entries(e).filter(function(a) {
            var n = ja(a, 2)
              , o = n[0]
              , i = n[1]
              , l = true;
            return Ai(o) || (console.warn('Skipped "'.concat(o, '" because it is not a valid MIME type. Check https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types for a list of valid MIME types.')),
            l = false),
            (!Array.isArray(i) || !i.every(ki)) && (console.warn('Skipped "'.concat(o, '" because an invalid file extension was provided.')),
            l = false),
            l;
        }).reduce(function(a, n) {
            var o = ja(n, 2)
              , i = o[0]
              , l = o[1];
            return _o(_o({}, a), {}, wi({}, i, l));
        }, {});
        return [{
            description: "Files",
            accept: t
        }];
    }
    return e;
}
function Nc(e) {
    if (Xt(e))
        return Object.entries(e).reduce(function(t, a) {
            var n = ja(a, 2)
              , o = n[0]
              , i = n[1];
            return [].concat(jo(t), [o], jo(i));
        }, []).filter(function(t) {
            return Ai(t) || ki(t);
        }).join(",");
}
function zc(e) {
    return e instanceof DOMException && (e.name === "AbortError" || e.code === e.ABORT_ERR);
}
function Vc(e) {
    return e instanceof DOMException && (e.name === "SecurityError" || e.code === e.SECURITY_ERR);
}
function Ai(e) {
    return e === "audio/*" || e === "video/*" || e === "image/*" || e === "text/*" || e === "application/*" || /\w+\/[-+.\w]+/g.test(e);
}
function ki(e) {
    return /^.*\.[\w]+$/.test(e);
}
var Hc = ["children"]
  , Uc = ["open"]
  , qc = ["refKey", "role", "onKeyDown", "onFocus", "onBlur", "onClick", "onDragEnter", "onDragOver", "onDragLeave", "onDrop"]
  , Wc = ["refKey", "onChange", "onClick"];
function Yc(e) {
    return Xc(e) || Kc(e) || Ti(e) || Gc();
}
function Gc() {
    throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`);
}
function Kc(e) {
    if (typeof Symbol < "u" && e[Symbol.iterator] != null || e["@@iterator"] != null)
        return Array.from(e);
}
function Xc(e) {
    if (Array.isArray(e))
        return kn(e);
}
function fn(e, t) {
    return Jc(e) || Zc(e, t) || Ti(e, t) || Qc();
}
function Qc() {
    throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`);
}
function Ti(e, t) {
    if (e) {
        if (typeof e == "string")
            return kn(e, t);
        var a = Object.prototype.toString.call(e).slice(8, -1);
        if (a === "Object" && e.constructor && (a = e.constructor.name),
        a === "Map" || a === "Set")
            return Array.from(e);
        if (a === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(a))
            return kn(e, t);
    }
}
function kn(e, t) {
    (t == null || t > e.length) && (t = e.length);
    for (var a = 0, n = new Array(t); a < t; a++)
        n[a] = e[a];
    return n;
}
function Zc(e, t) {
    var a = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
    if (a != null) {
        var n = [], o = true, i = false, l, s;
        try {
            for (a = a.call(e); !(o = (l = a.next()).done) && (n.push(l.value),
            !(t && n.length === t)); o = true)
                ;
        } catch (c) {
            i = true,
            s = c;
        } finally {
            try {
                !o && a.return != null && a.return();
            } finally {
                if (i)
                    throw s;
            }
        }
        return n;
    }
}
function Jc(e) {
    if (Array.isArray(e))
        return e;
}
function Lo(e, t) {
    var a = Object.keys(e);
    if (Object.getOwnPropertySymbols) {
        var n = Object.getOwnPropertySymbols(e);
        t && (n = n.filter(function(o) {
            return Object.getOwnPropertyDescriptor(e, o).enumerable;
        })),
        a.push.apply(a, n);
    }
    return a;
}
function Ce(e) {
    for (var t = 1; t < arguments.length; t++) {
        var a = arguments[t] != null ? arguments[t] : {};
        t % 2 ? Lo(Object(a), true).forEach(function(n) {
            Tn(e, n, a[n]);
        }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(a)) : Lo(Object(a)).forEach(function(n) {
            Object.defineProperty(e, n, Object.getOwnPropertyDescriptor(a, n));
        });
    }
    return e;
}
function Tn(e, t, a) {
    return t in e ? Object.defineProperty(e, t, {
        value: a,
        enumerable: true,
        configurable: true,
        writable: true
    }) : e[t] = a,
    e;
}
function tn(e, t) {
    if (e == null)
        return {};
    var a = ep(e, t), n, o;
    if (Object.getOwnPropertySymbols) {
        var i = Object.getOwnPropertySymbols(e);
        for (o = 0; o < i.length; o++)
            n = i[o],
            !(t.indexOf(n) >= 0) && Object.prototype.propertyIsEnumerable.call(e, n) && (a[n] = e[n]);
    }
    return a;
}
function ep(e, t) {
    if (e == null)
        return {};
    var a = {}, n = Object.keys(e), o, i;
    for (i = 0; i < n.length; i++)
        o = n[i],
        !(t.indexOf(o) >= 0) && (a[o] = e[o]);
    return a;
}
var ln = h.forwardRef(function(e, t) {
    var a = e.children
      , n = tn(e, Hc)
      , o = tp(n)
      , i = o.open
      , l = tn(o, Uc);
    return h.useImperativeHandle(t, function() {
        return {
            open: i
        };
    }, [i]),
    g.createElement(h.Fragment, null, a(Ce(Ce({}, l), {}, {
        open: i
    })));
});
ln.displayName = "Dropzone";
var Pi = {
    disabled: false,
    getFilesFromEvent: pc,
    maxSize: 1 / 0,
    minSize: 0,
    multiple: true,
    maxFiles: 0,
    preventDropOnDocument: true,
    noClick: false,
    noKeyboard: false,
    noDrag: false,
    noDragEventsBubbling: false,
    validator: null,
    useFsAccessApi: false,
    autoFocus: false
};
ln.defaultProps = Pi;
ln.propTypes = {
    children: de.func,
    accept: de.objectOf(de.arrayOf(de.string)),
    multiple: de.bool,
    preventDropOnDocument: de.bool,
    noClick: de.bool,
    noKeyboard: de.bool,
    noDrag: de.bool,
    noDragEventsBubbling: de.bool,
    minSize: de.number,
    maxSize: de.number,
    maxFiles: de.number,
    disabled: de.bool,
    getFilesFromEvent: de.func,
    onFileDialogCancel: de.func,
    onFileDialogOpen: de.func,
    useFsAccessApi: de.bool,
    autoFocus: de.bool,
    onDragEnter: de.func,
    onDragLeave: de.func,
    onDragOver: de.func,
    onDrop: de.func,
    onDropAccepted: de.func,
    onDropRejected: de.func,
    onError: de.func,
    validator: de.func
};
var Pn = {
    isFocused: false,
    isFileDialogActive: false,
    isDragActive: false,
    isDragAccept: false,
    isDragReject: false,
    acceptedFiles: [],
    fileRejections: []
};
function tp() {
    var e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {}
      , t = Ce(Ce({}, Pi), e)
      , a = t.accept
      , n = t.disabled
      , o = t.getFilesFromEvent
      , i = t.maxSize
      , l = t.minSize
      , s = t.multiple
      , c = t.maxFiles
      , d = t.onDragEnter
      , p = t.onDragLeave
      , u = t.onDragOver
      , f = t.onDrop
      , m = t.onDropAccepted
      , b = t.onDropRejected
      , y = t.onFileDialogCancel
      , w = t.onFileDialogOpen
      , E = t.useFsAccessApi
      , A = t.autoFocus
      , C = t.preventDropOnDocument
      , S = t.noClick
      , T = t.noKeyboard
      , P = t.noDrag
      , _ = t.noDragEventsBubbling
      , U = t.onError
      , L = t.validator
      , $ = h.useMemo(function() {
        return Nc(a);
    }, [a])
      , q = h.useMemo(function() {
        return Bc(a);
    }, [a])
      , M = h.useMemo(function() {
        return typeof w == "function" ? w : Mo;
    }, [w])
      , G = h.useMemo(function() {
        return typeof y == "function" ? y : Mo;
    }, [y])
      , H = h.useRef(null)
      , Z = h.useRef(null)
      , W = h.useReducer(ap, Pn)
      , le = fn(W, 2)
      , Ee = le[0]
      , ne = le[1]
      , Ae = Ee.isFocused
      , pe = Ee.isFileDialogActive
      , se = h.useRef(typeof window < "u" && window.isSecureContext && E && $c())
      , Ge = function() {
        !se.current && pe && setTimeout(function() {
            if (Z.current) {
                var I = Z.current.files;
                I.length || (ne({
                    type: "closeDialog"
                }),
                G());
            }
        }, 300);
    };
    h.useEffect(function() {
        return window.addEventListener("focus", Ge, false),
        function() {
            window.removeEventListener("focus", Ge, false);
        }
        ;
    }, [Z, pe, G, se]);
    var ye = h.useRef([])
      , ce = function(I) {
        H.current && H.current.contains(I.target) || (I.preventDefault(),
        ye.current = []);
    };
    h.useEffect(function() {
        return C && (document.addEventListener("dragover", Fo, false),
        document.addEventListener("drop", ce, false)),
        function() {
            C && (document.removeEventListener("dragover", Fo),
            document.removeEventListener("drop", ce));
        }
        ;
    }, [H, C]),
    h.useEffect(function() {
        return !n && A && H.current && H.current.focus(),
        function() {}
        ;
    }, [H, A, n]);
    var D = h.useCallback(function(k) {
        U ? U(k) : console.error(k);
    }, [U])
      , he = h.useCallback(function(k) {
        k.preventDefault(),
        k.persist(),
        Y(k),
        ye.current = [].concat(Yc(ye.current), [k.target]),
        qa(k) && Promise.resolve(o(k)).then(function(I) {
            if (!(en(k) && !_)) {
                var oe = I.length
                  , fe = oe > 0 && Dc({
                    files: I,
                    accept: $,
                    minSize: l,
                    maxSize: i,
                    multiple: s,
                    maxFiles: c,
                    validator: L
                })
                  , ve = oe > 0 && !fe;
                ne({
                    isDragAccept: fe,
                    isDragReject: ve,
                    isDragActive: true,
                    type: "setDraggedFiles"
                }),
                d && d(k);
            }
        }).catch(function(I) {
            return D(I);
        });
    }, [o, d, D, _, $, l, i, s, c, L])
      , Le = h.useCallback(function(k) {
        k.preventDefault(),
        k.persist(),
        Y(k);
        var I = qa(k);
        if (I && k.dataTransfer)
            try {
                k.dataTransfer.dropEffect = "copy";
            } catch {}
        return I && u && u(k),
        false;
    }, [u, _])
      , ue = h.useCallback(function(k) {
        k.preventDefault(),
        k.persist(),
        Y(k);
        var I = ye.current.filter(function(fe) {
            return H.current && H.current.contains(fe);
        })
          , oe = I.indexOf(k.target);
        oe !== -1 && I.splice(oe, 1),
        ye.current = I,
        !(I.length > 0) && (ne({
            type: "setDraggedFiles",
            isDragActive: false,
            isDragAccept: false,
            isDragReject: false
        }),
        qa(k) && p && p(k));
    }, [H, p, _])
      , ke = h.useCallback(function(k, I) {
        var oe = []
          , fe = [];
        k.forEach(function(ve) {
            var st = Si(ve, $)
              , we = fn(st, 2)
              , re = we[0]
              , Be = we[1]
              , Et = Ei(ve, l, i)
              , $e = fn(Et, 2)
              , ct = $e[0]
              , Xe = $e[1]
              , He = L ? L(ve) : null;
            if (re && ct && !He)
                oe.push(ve);
            else {
                var At = [Be, Xe];
                He && (At = At.concat(He)),
                fe.push({
                    file: ve,
                    errors: At.filter(function(pt) {
                        return pt;
                    })
                });
            }
        }),
        (!s && oe.length > 1 || s && c >= 1 && oe.length > c) && (oe.forEach(function(ve) {
            fe.push({
                file: ve,
                errors: [Rc]
            });
        }),
        oe.splice(0)),
        ne({
            acceptedFiles: oe,
            fileRejections: fe,
            isDragReject: fe.length > 0,
            type: "setFiles"
        }),
        f && f(oe, fe, I),
        fe.length > 0 && b && b(fe, I),
        oe.length > 0 && m && m(oe, I);
    }, [ne, s, $, l, i, c, f, m, b, L])
      , qe = h.useCallback(function(k) {
        k.preventDefault(),
        k.persist(),
        Y(k),
        ye.current = [],
        qa(k) && Promise.resolve(o(k)).then(function(I) {
            en(k) && !_ || ke(I, k);
        }).catch(function(I) {
            return D(I);
        }),
        ne({
            type: "reset"
        });
    }, [o, ke, D, _])
      , Qe = h.useCallback(function() {
        if (se.current) {
            ne({
                type: "openDialog"
            }),
            M();
            var k = {
                multiple: s,
                types: q
            };
            window.showOpenFilePicker(k).then(function(I) {
                return o(I);
            }).then(function(I) {
                ke(I, null),
                ne({
                    type: "closeDialog"
                });
            }).catch(function(I) {
                zc(I) ? (G(I),
                ne({
                    type: "closeDialog"
                })) : Vc(I) ? (se.current = false,
                Z.current ? (Z.current.value = null,
                Z.current.click()) : D(new Error("Cannot open the file picker because the https://developer.mozilla.org/en-US/docs/Web/API/File_System_Access_API is not supported and no <input> was provided."))) : D(I);
            });
            return;
        }
        Z.current && (ne({
            type: "openDialog"
        }),
        M(),
        Z.current.value = null,
        Z.current.click());
    }, [ne, M, G, E, ke, D, q, s])
      , Oe = h.useCallback(function(k) {
        !H.current || !H.current.isEqualNode(k.target) || (k.key === " " || k.key === "Enter" || k.keyCode === 32 || k.keyCode === 13) && (k.preventDefault(),
        Qe());
    }, [H, Qe])
      , rt = h.useCallback(function() {
        ne({
            type: "focus"
        });
    }, [])
      , St = h.useCallback(function() {
        ne({
            type: "blur"
        });
    }, [])
      , me = h.useCallback(function() {
        S || (Mc() ? setTimeout(Qe, 0) : Qe());
    }, [S, Qe])
      , Te = function(I) {
        return n ? null : I;
    }
      , R = function(I) {
        return T ? null : Te(I);
    }
      , Me = function(I) {
        return P ? null : Te(I);
    }
      , Y = function(I) {
        _ && I.stopPropagation();
    }
      , _e = h.useMemo(function() {
        return function() {
            var k = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {}
              , I = k.refKey
              , oe = I === void 0 ? "ref" : I
              , fe = k.role
              , ve = k.onKeyDown
              , st = k.onFocus
              , we = k.onBlur
              , re = k.onClick
              , Be = k.onDragEnter
              , Et = k.onDragOver
              , $e = k.onDragLeave
              , ct = k.onDrop
              , Xe = tn(k, qc);
            return Ce(Ce(Tn({
                onKeyDown: R(Pt(ve, Oe)),
                onFocus: R(Pt(st, rt)),
                onBlur: R(Pt(we, St)),
                onClick: Te(Pt(re, me)),
                onDragEnter: Me(Pt(Be, he)),
                onDragOver: Me(Pt(Et, Le)),
                onDragLeave: Me(Pt($e, ue)),
                onDrop: Me(Pt(ct, qe)),
                role: typeof fe == "string" && fe !== "" ? fe : "presentation"
            }, oe, H), !n && !T ? {
                tabIndex: 0
            } : {}), Xe);
        }
        ;
    }, [H, Oe, rt, St, me, he, Le, ue, qe, T, P, n])
      , Ke = h.useCallback(function(k) {
        k.stopPropagation();
    }, [])
      , lt = h.useMemo(function() {
        return function() {
            var k = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {}
              , I = k.refKey
              , oe = I === void 0 ? "ref" : I
              , fe = k.onChange
              , ve = k.onClick
              , st = tn(k, Wc)
              , we = Tn({
                accept: $,
                multiple: s,
                type: "file",
                style: {
                    border: 0,
                    clip: "rect(0, 0, 0, 0)",
                    clipPath: "inset(50%)",
                    height: "1px",
                    margin: "0 -1px -1px 0",
                    overflow: "hidden",
                    padding: 0,
                    position: "absolute",
                    width: "1px",
                    whiteSpace: "nowrap"
                },
                onChange: Te(Pt(fe, qe)),
                onClick: Te(Pt(ve, Ke)),
                tabIndex: -1
            }, oe, Z);
            return Ce(Ce({}, we), st);
        }
        ;
    }, [Z, a, s, qe, n]);
    return Ce(Ce({}, Ee), {}, {
        isFocused: Ae && !n,
        getRootProps: _e,
        getInputProps: lt,
        rootRef: H,
        inputRef: Z,
        open: Te(Qe)
    });
}
function ap(e, t) {
    switch (t.type) {
    case "focus":
        return Ce(Ce({}, e), {}, {
            isFocused: true
        });
    case "blur":
        return Ce(Ce({}, e), {}, {
            isFocused: false
        });
    case "openDialog":
        return Ce(Ce({}, Pn), {}, {
            isFileDialogActive: true
        });
    case "closeDialog":
        return Ce(Ce({}, e), {}, {
            isFileDialogActive: false
        });
    case "setDraggedFiles":
        return Ce(Ce({}, e), {}, {
            isDragActive: t.isDragActive,
            isDragAccept: t.isDragAccept,
            isDragReject: t.isDragReject
        });
    case "setFiles":
        return Ce(Ce({}, e), {}, {
            acceptedFiles: t.acceptedFiles,
            fileRejections: t.fileRejections,
            isDragReject: t.isDragReject
        });
    case "reset":
        return Ce({}, Pn);
    default:
        return e;
    }
}
function Mo() {}
const np = ({onClick: e, dataTest: t="close_button", color: a, className: n, disabled: o}) => {
    const i = Se("m-1", "ms-0", n)
      , l = g.useCallback( () => {
        e == null ? void 0 : e();
    }
    , [e]);
    return r.jsx($n, {
        "data-testid": t,
        id: t,
        onClick: l,
        className: i,
        color: a,
        disabled: o,
        children: "\u0417\u0430\u043A\u0440\u044B\u0442\u044C"
    });
}
  , Oi = ({dataTest: e="information_modal", showModal: t=false, setShowModal: a, title: n, content: o}) => {
    const i = () => {
        a == null ? void 0 : a(!t);
    }
    ;
    return r.jsxs(ar, {
        "data-testid": e,
        open: t,
        children: [r.jsx(nr, {
            variant: "h2",
            children: n
        }), r.jsx(or, {
            children: o
        }), r.jsx(ir, {
            children: r.jsx(np, {
                onClick: i,
                className: "w-25"
            })
        })]
    });
}
  , op = ({dataTest: e="question_icon", variant: t="thin", title: a, content: n, sx: o}) => {
    const [i,l] = g.useState(false)
      , s = g.useCallback( (c) => {
        c.preventDefault(),
        l(!i);
    }
    , [i]);
    return r.jsxs(K, {
        "data-testid": e,
        children: [r.jsx(Wt, {
            dataTest: "question_button",
            icon: t === "thin" ? je.questionCircleThin : je.questionCircle,
            onClick: s,
            sx: o
        }), r.jsx(Oi, {
            showModal: i,
            setShowModal: l,
            title: a,
            content: n
        })]
    });
}
  , Jn = ({dataTest: e="delete_icon", onClick: t, color: a="error", sx: n, variant: o, isDisabled: i}) => r.jsx(Wt, {
    dataTest: e,
    icon: o === "thin" ? je.trashCanThin : je.trashCan,
    onClick: t,
    color: a,
    sx: n,
    isDisabled: i
})
  , ip = {
    selectAll: "(Select All)",
    selectAllSearchResults: "(Select All Search Results)",
    addCurrentSelectionToFilter: "Add current selection to filter",
    searchOoo: "Search...",
    blanks: "(Blanks)",
    noMatches: "No matches",
    filterOoo: "Filter...",
    equals: "Equals",
    notEqual: "Does not equal",
    blank: "Blank",
    notBlank: "Not blank",
    empty: "Choose one",
    lessThan: "Less than",
    greaterThan: "Greater than",
    lessThanOrEqual: "Less than or equal to",
    greaterThanOrEqual: "Greater than or equal to",
    inRange: "Between",
    inRangeStart: "From",
    inRangeEnd: "To",
    contains: "Contains",
    notContains: "Does not contain",
    startsWith: "Begins with",
    endsWith: "Ends with",
    dateFormatOoo: "yyyy-mm-dd",
    before: "Before",
    after: "After",
    andCondition: "AND",
    orCondition: "OR",
    applyFilter: "\u041F\u0440\u0438\u043C\u0435\u043D\u0438\u0442\u044C",
    resetFilter: "\u0421\u0431\u0440\u043E\u0441\u0438\u0442\u044C",
    clearFilter: "\u041E\u0447\u0438\u0441\u0442\u0438\u0442\u044C",
    cancelFilter: "\u041E\u0442\u043C\u0435\u043D\u0438\u0442\u044C",
    textFilter: "Text Filter",
    numberFilter: "Number Filter",
    dateFilter: "Date Filter",
    setFilter: "Set Filter",
    groupFilterSelect: "Select field:",
    advancedFilterContains: "contains",
    advancedFilterNotContains: "does not contain",
    advancedFilterTextEquals: "equals",
    advancedFilterTextNotEqual: "does not equal",
    advancedFilterStartsWith: "begins with",
    advancedFilterEndsWith: "ends with",
    advancedFilterBlank: "is blank",
    advancedFilterNotBlank: "is not blank",
    advancedFilterEquals: "=",
    advancedFilterNotEqual: "!=",
    advancedFilterGreaterThan: ">",
    advancedFilterGreaterThanOrEqual: ">=",
    advancedFilterLessThan: "<",
    advancedFilterLessThanOrEqual: "<=",
    advancedFilterTrue: "is true",
    advancedFilterFalse: "is false",
    advancedFilterAnd: "AND",
    advancedFilterOr: "OR",
    advancedFilterApply: "Apply",
    advancedFilterBuilder: "Builder",
    advancedFilterValidationMissingColumn: "Column is missing",
    advancedFilterValidationMissingOption: "Option is missing",
    advancedFilterValidationMissingValue: "Value is missing",
    advancedFilterValidationInvalidColumn: "Column not found",
    advancedFilterValidationInvalidOption: "Option not found",
    advancedFilterValidationMissingQuote: "Value is missing an end quote",
    advancedFilterValidationNotANumber: "Value is not a number",
    advancedFilterValidationInvalidDate: "Value is not a valid date",
    advancedFilterValidationMissingCondition: "Condition is missing",
    advancedFilterValidationJoinOperatorMismatch: "Join operators within a condition must be the same",
    advancedFilterValidationInvalidJoinOperator: "Join operator not found",
    advancedFilterValidationMissingEndBracket: "Missing end bracket",
    advancedFilterValidationExtraEndBracket: "Too many end brackets",
    advancedFilterValidationMessage: "Expression has an error. ${variable} - ${variable}.",
    advancedFilterValidationMessageAtEnd: "Expression has an error. ${variable} at end of expression.",
    advancedFilterBuilderTitle: "Advanced Filter",
    advancedFilterBuilderApply: "Apply",
    advancedFilterBuilderCancel: "Cancel",
    advancedFilterBuilderAddButtonTooltip: "Add Filter or Group",
    advancedFilterBuilderRemoveButtonTooltip: "Remove",
    advancedFilterBuilderMoveUpButtonTooltip: "Move Up",
    advancedFilterBuilderMoveDownButtonTooltip: "Move Down",
    advancedFilterBuilderAddJoin: "Add Group",
    advancedFilterBuilderAddCondition: "Add Filter",
    advancedFilterBuilderSelectColumn: "Select a column",
    advancedFilterBuilderSelectOption: "Select an option",
    advancedFilterBuilderEnterValue: "Enter a value...",
    advancedFilterBuilderValidationAlreadyApplied: "Current filter already applied.",
    advancedFilterBuilderValidationIncomplete: "Not all conditions are complete.",
    advancedFilterBuilderValidationSelectColumn: "Must select a column.",
    advancedFilterBuilderValidationSelectOption: "Must select an option.",
    advancedFilterBuilderValidationEnterValue: "Must enter a value.",
    columns: "Columns",
    filters: "Filters",
    pivotMode: "Pivot Mode",
    groups: "Row Groups",
    rowGroupColumnsEmptyMessage: "Drag here to set row groups",
    values: "Values",
    valueColumnsEmptyMessage: "Drag here to aggregate",
    pivots: "Column Labels",
    pivotColumnsEmptyMessage: "Drag here to set column labels",
    group: "Group",
    rowDragRow: "row",
    rowDragRows: "rows",
    loadingOoo: "\u0417\u0430\u0433\u0440\u0443\u0436\u0430\u044E \u0434\u0430\u043D\u043D\u044B\u0435 \u0441 \u0441\u0435\u0440\u0432\u0435\u0440\u0430...",
    loadingError: "ERR",
    noRowsToShow: "\u041D\u0438\u0447\u0435\u0433\u043E \u043D\u0435 \u043D\u0430\u0439\u0434\u0435\u043D\u043E",
    enabled: "Enabled",
    pinColumn: "\u0417\u0430\u043A\u0440\u0435\u043F\u0438\u0442\u044C \u0441\u0442\u043E\u043B\u0431\u0435\u0446",
    pinLeft: "\u0417\u0430\u043A\u0440\u0435\u043F\u0438\u0442\u044C \u0441\u043B\u0435\u0432\u0430",
    pinRight: "\u0417\u0430\u043A\u0440\u0435\u043F\u0438\u0442\u044C \u0441\u043F\u0440\u0430\u0432\u0430",
    noPin: "\u0421\u043D\u044F\u0442\u044C \u0437\u0430\u043A\u0440\u0435\u043F\u043B\u0435\u043D\u0438\u0435",
    valueAggregation: "Value Aggregation",
    noAggregation: "None",
    autosizeThisColumn: "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0447\u0435\u0441\u043A\u043E\u0435 \u0438\u0437\u043C\u0435\u043D\u0435\u043D\u0438\u0435 \u0440\u0430\u0437\u043C\u0435\u0440\u0430 \u0441\u0442\u043E\u043B\u0431\u0446\u0430",
    autosizeAllColumns: "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0447\u0435\u0441\u043A\u043E\u0435 \u0438\u0437\u043C\u0435\u043D\u0435\u043D\u0438\u0435 \u0440\u0430\u0437\u043C\u0435\u0440\u0430 \u0432\u0441\u0435\u0445 \u0441\u0442\u043E\u043B\u0431\u0446\u043E\u0432",
    groupBy: "Group by",
    ungroupBy: "Un-Group by",
    ungroupAll: "Un-Group All",
    addToValues: "Add ${variable} to values",
    removeFromValues: "Remove ${variable} from values",
    addToLabels: "Add ${variable} to labels",
    removeFromLabels: "Remove ${variable} from labels",
    resetColumns: "\u0421\u0431\u0440\u043E\u0441\u0438\u0442\u044C \u0441\u0442\u043E\u043B\u0431\u0446\u044B \u043A \u0441\u043E\u0441\u0442\u043E\u044F\u043D\u0438\u044E \u043F\u043E \u0443\u043C\u043E\u043B\u0447\u0430\u043D\u0438\u044E",
    expandAll: "Expand All Row Groups",
    collapseAll: "Close All Row Groups",
    copy: "Copy",
    ctrlC: "Ctrl+C",
    ctrlX: "Ctrl+X",
    copyWithHeaders: "Copy With Headers",
    copyWithGroupHeaders: "Copy with Group Headers",
    cut: "Cut",
    paste: "Paste",
    ctrlV: "Ctrl+V",
    export: "Export",
    csvExport: "CSV Export",
    excelExport: "Excel Export",
    columnFilter: "Column Filter",
    columnChooser: "\u0412\u044B\u0431\u0440\u0430\u0442\u044C \u0441\u0442\u043E\u043B\u0431\u0446\u044B",
    chooseColumns: "\u0412\u044B\u0431\u0440\u0430\u0442\u044C \u0441\u0442\u043E\u043B\u0431\u0446\u044B",
    sortAscending: "\u0421\u043E\u0440\u0442\u0438\u0440\u043E\u0432\u043A\u0430 \u043F\u043E \u0432\u043E\u0437\u0440\u0430\u0441\u0442\u0430\u043D\u0438\u044E",
    sortDescending: "\u0421\u043E\u0440\u0442\u0438\u0440\u043E\u0432\u043A\u0430 \u043F\u043E \u0443\u0431\u044B\u0432\u0430\u043D\u0438\u044E",
    sortUnSort: "\u041E\u0447\u0438\u0441\u0442\u0438\u0442\u044C \u0441\u043E\u0440\u0442\u0438\u0440\u043E\u0432\u043A\u0443",
    sum: "Sum",
    first: "First",
    last: "Last",
    min: "Min",
    max: "Max",
    none: "None",
    count: "Count",
    avg: "Average",
    filteredRows: "Filtered",
    selectedRows: "Selected",
    totalRows: "Total Rows",
    totalAndFilteredRows: "Rows",
    more: "More",
    to: "\u043F\u043E",
    of: "\u0438\u0437",
    page: "\u0421\u0442\u0440\u0430\u043D\u0438\u0446\u0430",
    pageLastRowUnknown: "?",
    nextPage: "Next Page",
    lastPage: "Last Page",
    firstPage: "First Page",
    previousPage: "Previous Page",
    pageSizeSelectorLabel: "\u0421\u0442\u0440\u043E\u043A \u043D\u0430 \u0441\u0442\u0440\u0430\u043D\u0438\u0446\u0435:",
    footerTotal: "Total",
    pivotColumnGroupTotals: "Total",
    pivotChartAndPivotMode: "Pivot Chart & Pivot Mode",
    pivotChart: "Pivot Chart",
    chartRange: "Chart Range",
    columnChart: "Column",
    groupedColumn: "Grouped",
    stackedColumn: "Stacked",
    normalizedColumn: "100% Stacked",
    barChart: "Bar",
    groupedBar: "Grouped",
    stackedBar: "Stacked",
    normalizedBar: "100% Stacked",
    pieChart: "Pie",
    pie: "Pie",
    donut: "Donut",
    line: "Line",
    stackedLine: "Stacked",
    normalizedLine: "100% Stacked",
    xyChart: "X Y (Scatter)",
    scatter: "Scatter",
    bubble: "Bubble",
    areaChart: "Area",
    area: "Area",
    stackedArea: "Stacked",
    normalizedArea: "100% Stacked",
    histogramChart: "Histogram",
    polarChart: "Polar",
    radarLine: "Radar Line",
    radarArea: "Radar Area",
    nightingale: "Nightingale",
    radialColumn: "Radial Column",
    radialBar: "Radial Bar",
    statisticalChart: "Statistical",
    boxPlot: "Box Plot",
    rangeBar: "Range Bar",
    rangeArea: "Range Area",
    hierarchicalChart: "Hierarchical",
    treemap: "Treemap",
    sunburst: "Sunburst",
    specializedChart: "Specialized",
    waterfall: "Waterfall",
    heatmap: "Heatmap",
    combinationChart: "Combination",
    columnLineCombo: "Column & Line",
    AreaColumnCombo: "Area & Column",
    pivotChartTitle: "Pivot Chart",
    rangeChartTitle: "Range Chart",
    settings: "Chart",
    data: "Set Up",
    format: "Customize",
    categories: "Categories",
    defaultCategory: "(None)",
    series: "Series",
    switchCategorySeries: "Switch Category / Series",
    categoryValues: "Category Values",
    seriesLabels: "Series Labels",
    aggregate: "Aggregate",
    xyValues: "X Y Values",
    paired: "Paired Mode",
    axis: "Axis",
    xAxis: "Horizontal Axis",
    yAxis: "Vertical Axis",
    polarAxis: "Polar Axis",
    radiusAxis: "Radius Axis",
    navigator: "Navigator",
    zoom: "Zoom",
    animation: "Animation",
    crosshair: "Crosshair",
    color: "Color",
    thickness: "Thickness",
    preferredLength: "Preferred Length",
    xType: "X Type",
    axisType: "Axis Type",
    automatic: "Automatic",
    category: "Category",
    number: "Number",
    time: "Time",
    timeFormat: "Time Format",
    autoRotate: "Auto Rotate",
    labelRotation: "Rotation",
    circle: "Circle",
    polygon: "Polygon",
    square: "Square",
    cross: "Cross",
    diamond: "Diamond",
    plus: "Plus",
    triangle: "Triangle",
    heart: "Heart",
    orientation: "Orientation",
    fixed: "Fixed",
    parallel: "Parallel",
    perpendicular: "Perpendicular",
    radiusAxisPosition: "Position",
    ticks: "Ticks",
    gridLines: "Grid Lines",
    width: "Width",
    height: "Height",
    length: "Length",
    padding: "Padding",
    spacing: "Spacing",
    chartStyle: "Chart Style",
    title: "Title",
    chartTitles: "Titles",
    chartTitle: "Chart Title",
    chartSubtitle: "Subtitle",
    horizontalAxisTitle: "Horizontal Axis Title",
    verticalAxisTitle: "Vertical Axis Title",
    polarAxisTitle: "Polar Axis Title",
    titlePlaceholder: "Chart Title",
    background: "Background",
    font: "Font",
    weight: "Weight",
    top: "Top",
    right: "Right",
    bottom: "Bottom",
    left: "Left",
    labels: "Labels",
    calloutLabels: "Callout Labels",
    sectorLabels: "Sector Labels",
    positionRatio: "Position Ratio",
    size: "Size",
    shape: "Shape",
    minSize: "Minimum Size",
    maxSize: "Maximum Size",
    legend: "Legend",
    position: "Position",
    markerSize: "Marker Size",
    markerStroke: "Marker Stroke",
    markerPadding: "Marker Padding",
    itemSpacing: "Item Spacing",
    itemPaddingX: "Item Padding X",
    itemPaddingY: "Item Padding Y",
    layoutHorizontalSpacing: "Horizontal Spacing",
    layoutVerticalSpacing: "Vertical Spacing",
    strokeWidth: "Stroke Width",
    offset: "Offset",
    offsets: "Offsets",
    tooltips: "Tooltips",
    callout: "Callout",
    markers: "Markers",
    shadow: "Shadow",
    blur: "Blur",
    xOffset: "X Offset",
    yOffset: "Y Offset",
    lineWidth: "Line Width",
    lineDash: "Line Dash",
    lineDashOffset: "Dash Offset",
    scrollingZoom: "Scrolling",
    scrollingStep: "Scrolling Step",
    selectingZoom: "Selecting",
    durationMillis: "Duration (ms)",
    crosshairLabel: "Label",
    crosshairSnap: "Snap to Node",
    normal: "Normal",
    bold: "Bold",
    italic: "Italic",
    boldItalic: "Bold Italic",
    predefined: "Predefined",
    fillOpacity: "Fill Opacity",
    strokeColor: "Line Color",
    strokeOpacity: "Line Opacity",
    miniChart: "Mini-Chart",
    histogramBinCount: "Bin count",
    connectorLine: "Connector Line",
    seriesItems: "Series Items",
    seriesItemType: "Item Type",
    seriesItemPositive: "Positive",
    seriesItemNegative: "Negative",
    seriesItemLabels: "Item Labels",
    columnGroup: "Column",
    barGroup: "Bar",
    pieGroup: "Pie",
    lineGroup: "Line",
    scatterGroup: "X Y (Scatter)",
    areaGroup: "Area",
    polarGroup: "Polar",
    statisticalGroup: "Statistical",
    hierarchicalGroup: "Hierarchical",
    specializedGroup: "Specialized",
    combinationGroup: "Combination",
    groupedColumnTooltip: "Grouped",
    stackedColumnTooltip: "Stacked",
    normalizedColumnTooltip: "100% Stacked",
    groupedBarTooltip: "Grouped",
    stackedBarTooltip: "Stacked",
    normalizedBarTooltip: "100% Stacked",
    pieTooltip: "Pie",
    donutTooltip: "Donut",
    lineTooltip: "Line",
    groupedAreaTooltip: "Area",
    stackedAreaTooltip: "Stacked",
    normalizedAreaTooltip: "100% Stacked",
    scatterTooltip: "Scatter",
    bubbleTooltip: "Bubble",
    histogramTooltip: "Histogram",
    radialColumnTooltip: "Radial Column",
    radialBarTooltip: "Radial Bar",
    radarLineTooltip: "Radar Line",
    radarAreaTooltip: "Radar Area",
    nightingaleTooltip: "Nightingale",
    rangeBarTooltip: "Range Bar",
    rangeAreaTooltip: "Range Area",
    boxPlotTooltip: "Box Plot",
    treemapTooltip: "Treemap",
    sunburstTooltip: "Sunburst",
    waterfallTooltip: "Waterfall",
    heatmapTooltip: "Heatmap",
    columnLineComboTooltip: "Column & Line",
    areaColumnComboTooltip: "Area & Column",
    customComboTooltip: "Custom Combination",
    innerRadius: "Inner Radius",
    startAngle: "Start Angle",
    endAngle: "End Angle",
    reverseDirection: "Reverse Direction",
    groupPadding: "Group Padding",
    seriesPadding: "Series Padding",
    tile: "Tile",
    whisker: "Whisker",
    cap: "Cap",
    capLengthRatio: "Length Ratio",
    labelPlacement: "Placement",
    inside: "Inside",
    outside: "Outside",
    noDataToChart: "No data available to be charted.",
    pivotChartRequiresPivotMode: "Pivot Chart requires Pivot Mode enabled.",
    chartSettingsToolbarTooltip: "Menu",
    chartLinkToolbarTooltip: "Linked to Grid",
    chartUnlinkToolbarTooltip: "Unlinked from Grid",
    chartDownloadToolbarTooltip: "Download Chart",
    chartMenuToolbarTooltip: "Menu",
    chartEdit: "Edit Chart",
    chartAdvancedSettings: "Advanced Settings",
    chartLink: "Link to Grid",
    chartUnlink: "Unlink from Grid",
    chartDownload: "Download Chart",
    histogramFrequency: "Frequency",
    seriesChartType: "Series Chart Type",
    seriesType: "Series Type",
    secondaryAxis: "Secondary Axis",
    seriesAdd: "Add a series",
    categoryAdd: "Add a category",
    bar: "Bar",
    column: "Column",
    histogram: "Histogram",
    advancedSettings: "Advanced Settings",
    direction: "Direction",
    horizontal: "Horizontal",
    vertical: "Vertical",
    seriesGroupType: "Group Type",
    groupedSeriesGroupType: "Grouped",
    stackedSeriesGroupType: "Stacked",
    normalizedSeriesGroupType: "100% Stacked",
    legendEnabled: "Enabled",
    invalidColor: "Color value is invalid",
    groupedColumnFull: "Grouped Column",
    stackedColumnFull: "Stacked Column",
    normalizedColumnFull: "100% Stacked Column",
    groupedBarFull: "Grouped Bar",
    stackedBarFull: "Stacked Bar",
    normalizedBarFull: "100% Stacked Bar",
    stackedAreaFull: "Stacked Area",
    normalizedAreaFull: "100% Stacked Area",
    customCombo: "Custom Combination",
    ariaAdvancedFilterBuilderItem: "${variable}. Level ${variable}. Press ENTER to edit",
    ariaAdvancedFilterBuilderItemValidation: "${variable}. Level ${variable}. ${variable} Press ENTER to edit",
    ariaAdvancedFilterBuilderList: "Advanced Filter Builder List",
    ariaAdvancedFilterBuilderFilterItem: "Filter Condition",
    ariaAdvancedFilterBuilderGroupItem: "Filter Group",
    ariaAdvancedFilterBuilderColumn: "Column",
    ariaAdvancedFilterBuilderOption: "Option",
    ariaAdvancedFilterBuilderValueP: "Value",
    ariaAdvancedFilterBuilderJoinOperator: "Join Operator",
    ariaAdvancedFilterInput: "Advanced Filter Input",
    ariaChecked: "checked",
    ariaColumn: "Column",
    ariaColumnGroup: "Column Group",
    ariaColumnFiltered: "Column Filtered",
    ariaColumnSelectAll: "Toggle All Columns Visibility",
    ariaDateFilterInput: "Date Filter Input",
    ariaDefaultListName: "List",
    ariaFilterColumnsInput: "Filter Columns Input",
    ariaFilterFromValue: "Filter from value",
    ariaFilterInput: "Filter Input",
    ariaFilterList: "Filter List",
    ariaFilterToValue: "Filter to value",
    ariaFilterValue: "Filter Value",
    ariaFilterMenuOpen: "Open Filter Menu",
    ariaFilteringOperator: "Filtering Operator",
    ariaHidden: "hidden",
    ariaIndeterminate: "indeterminate",
    ariaInputEditor: "Input Editor",
    ariaMenuColumn: "Press ALT DOWN to open column menu",
    ariaFilterColumn: "Press CTRL ENTER to open filter",
    ariaRowDeselect: "Press SPACE to deselect this row",
    ariaHeaderSelection: "Column with Header Selection",
    ariaRowSelectAll: "Press Space to toggle all rows selection",
    ariaRowToggleSelection: "Press Space to toggle row selection",
    ariaRowSelect: "Press SPACE to select this row",
    ariaRowSelectionDisabled: "Row Selection is disabled for this row",
    ariaSearch: "Search",
    ariaSortableColumn: "Press ENTER to sort",
    ariaToggleVisibility: "Press SPACE to toggle visibility",
    ariaToggleCellValue: "Press SPACE to toggle cell value",
    ariaUnchecked: "unchecked",
    ariaVisible: "visible",
    ariaSearchFilterValues: "Search filter values",
    ariaPageSizeSelectorLabel: "Page Size",
    ariaChartMenuClose: "Close Chart Edit Menu",
    ariaChartSelected: "Selected",
    ariaSkeletonCellLoadingFailed: "Row failed to load",
    ariaSkeletonCellLoading: "Row data is loading",
    ariaRowGroupDropZonePanelLabel: "Row Groups",
    ariaValuesDropZonePanelLabel: "Values",
    ariaPivotDropZonePanelLabel: "Column Labels",
    ariaDropZoneColumnComponentDescription: "Press DELETE to remove",
    ariaDropZoneColumnValueItemDescription: "Press ENTER to change the aggregation type",
    ariaDropZoneColumnGroupItemDescription: "Press ENTER to sort",
    ariaDropZoneColumnComponentAggFuncSeparator: " of ",
    ariaDropZoneColumnComponentSortAscending: "ascending",
    ariaDropZoneColumnComponentSortDescending: "descending",
    ariaLabelDialog: "Dialog",
    ariaLabelColumnMenu: "Column Menu",
    ariaLabelColumnFilter: "Column Filter",
    ariaLabelCellEditor: "Cell Editor",
    ariaLabelSelectField: "Select Field",
    ariaLabelRichSelectField: "Rich Select Field",
    ariaLabelRichSelectToggleSelection: "Press SPACE to toggle selection",
    ariaLabelRichSelectDeselectAllItems: "Press DELETE to deselect all items",
    ariaLabelRichSelectDeleteSelection: "Press DELETE to deselect item",
    ariaLabelTooltip: "Tooltip",
    ariaLabelContextMenu: "Context Menu",
    ariaLabelSubMenu: "SubMenu",
    ariaLabelAggregationFunction: "Aggregation Function",
    ariaLabelAdvancedFilterAutocomplete: "Advanced Filter Autocomplete",
    ariaLabelAdvancedFilterBuilderAddField: "Advanced Filter Builder Add Field",
    ariaLabelAdvancedFilterBuilderColumnSelectField: "Advanced Filter Builder Column Select Field",
    ariaLabelAdvancedFilterBuilderOptionSelectField: "Advanced Filter Builder Option Select Field",
    ariaLabelAdvancedFilterBuilderJoinSelectField: "Advanced Filter Builder Join Operator Select Field",
    ariaColumnPanelList: "Column List",
    ariaFilterPanelList: "Filter List",
    thousandSeparator: ",",
    decimalSeparator: ".",
    true: "True",
    false: "False",
    invalidDate: "Invalid Date",
    invalidNumber: "Invalid Number",
    january: "January",
    february: "February",
    march: "March",
    april: "April",
    may: "May",
    june: "June",
    july: "July",
    august: "August",
    september: "September",
    october: "October",
    november: "November",
    december: "December",
    timeFormatSlashesDDMMYYYY: "DD/MM/YYYY",
    timeFormatSlashesMMDDYYYY: "MM/DD/YYYY",
    timeFormatSlashesDDMMYY: "DD/MM/YY",
    timeFormatSlashesMMDDYY: "MM/DD/YY",
    timeFormatDotsDDMYY: "DD.M.YY",
    timeFormatDotsMDDYY: "M.DD.YY",
    timeFormatDashesYYYYMMDD: "YYYY-MM-DD",
    timeFormatSpacesDDMMMMYYYY: "DD MMMM YYYY",
    timeFormatHHMMSS: "HH:MM:SS",
    timeFormatHHMMSSAmPm: "HH:MM:SS AM/PM"
}
  , ji = ({dataTest: e="file_type_icon", fileType: t, styles: a, sx: n}) => {
    const o = g.useMemo( () => {
        const [i,l] = t.split("/");
        switch (i) {
        case "image":
            return je.imageThin;
        case "video":
            return je.videoThin;
        case "audio":
            return je.audioThin;
        case "text":
            return l === "csv" ? je.csvThin : je.textThin;
        case "application":
            switch (l) {
            case "pdf":
                return je.pdfThin;
            case "vnd.ms-powerpoint":
            case "vnd.openxmlformats-officedocument.presentationml.presentation":
                return je.pptThin;
            case "msword":
            case "vnd.openxmlformats-officedocument.wordprocessingml.document":
                return je.wordThin;
            case "vnd.ms-excel":
            case "vnd.openxmlformats-officedocument.spreadsheetml.sheet":
                return je.excelThin;
            default:
                return je.fileThin;
            }
        default:
            return je.fileThin;
        }
    }
    , [t]);
    return r.jsx(Wt, {
        dataTest: e,
        icon: o,
        style: {
            fontSize: "1rem",
            cursor: "pointer",
            pointerEvents: "none",
            ...a
        },
        sx: n
    });
}
  , rp = ({dataTest: e="download_icon", onClick: t, sx: a, color: n, isDisabled: o}) => {
    const i = g.useCallback( (l) => {
        l.preventDefault(),
        t == null ? void 0 : t(l);
    }
    , [t]);
    return r.jsx(Wt, {
        dataTest: e,
        icon: je.downloadFileThin,
        onClick: i,
        sx: a,
        color: n,
        isDisabled: o
    });
}
  , lp = (e) => z.isEmpty(e)
  , $o = (e) => Array.isArray(e) ? e : [e]
  , sp = async (e, t) => {
    try {
        const a = await an({
            url: e,
            method: "get",
            contentType: "application/octet-stream",
            responseType: "blob"
        });
        return rr(a, t);
    } catch (a) {
        console.error(a);
    }
}
  , wa = (e) => {
    var _a2, _b, _c2, _d2, _e, _f, _g, _h, _i2, _j, _k, _l2, _m, _n2, _o2, _p2, _q, _r2, _s2, _t2;
    return {
        root: {
            position: "relative",
            width: "100%",
            minHeight: "250px",
            backgroundColor: (_b = (_a2 = e.palette) == null ? void 0 : _a2.background) == null ? void 0 : _b.paper,
            border: "dashed",
            borderColor: (_c2 = e.palette) == null ? void 0 : _c2.divider,
            borderRadius: (_d2 = e.shape) == null ? void 0 : _d2.borderRadius,
            boxSizing: "border-box",
            cursor: "pointer",
            overflow: "hidden",
            display: "flex"
        },
        active: {
            animation: "$progress 2s linear infinite !important",
            backgroundImage: `repeating-linear-gradient(-45deg, ${Na[200]}, ${Na[200]} 25px, ${Na[500]} 25px, ${Na[500]} 50px)`,
            backgroundSize: "150% 100%",
            border: "solid",
            borderColor: (_f = (_e = e.palette) == null ? void 0 : _e.primary) == null ? void 0 : _f.light
        },
        invalid: {
            backgroundImage: `repeating-linear-gradient(-45deg, ${(_h = (_g = e.palette) == null ? void 0 : _g.error) == null ? void 0 : _h.light}, ${(_j = (_i2 = e.palette) == null ? void 0 : _i2.error) == null ? void 0 : _j.light} 25px, ${(_l2 = (_k = e.palette) == null ? void 0 : _k.error) == null ? void 0 : _l2.dark} 25px, ${(_n2 = (_m = e.palette) == null ? void 0 : _m.error) == null ? void 0 : _n2.dark} 50px)`,
            borderColor: (_p2 = (_o2 = e.palette) == null ? void 0 : _o2.error) == null ? void 0 : _p2.main
        },
        textContainer: {
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
            width: "100%"
        },
        text: {
            marginBottom: (_q = e.spacing) == null ? void 0 : _q.call(e, 3),
            marginTop: (_r2 = e.spacing) == null ? void 0 : _r2.call(e, 3)
        },
        icon: {
            width: 51,
            height: 51,
            color: (_t2 = (_s2 = e.palette) == null ? void 0 : _s2.text) == null ? void 0 : _t2.primary
        }
    };
}
  , cp = ({dataTest: e="dropzone_content", getRootProps: t, getInputProps: a, isDragActive: n, isDragReject: o, dropzoneText: i, icon: l, inputProps: s, disableRejectionFeedback: c, dropZoneStyles: d={}, dropzoneParagraphStyles: p={}, maxFileSizeFormatted: u}) => r.jsxs(K, {
    "data-testid": e,
    children: [r.jsxs(K, {
        component: "div",
        ...t(),
        sx: [ (f) => ({
            ...wa(f).root,
            ...n && wa(f).active,
            ...!c && o && wa(f).invalid
        }), ...$o(d)],
        children: [r.jsx("input", {
            ...a(s),
            "data-testid": "file_input"
        }), r.jsxs(K, {
            sx: (f) => wa(f).textContainer,
            children: [r.jsx(F, {
                variant: "h5",
                component: "p",
                sx: [ (f) => ({
                    ...wa(f).text
                }), ...$o(p)],
                children: i
            }), r.jsx(Wt, {
                icon: l || je.uploadCloud,
                sx: {
                    fontSize: "2rem",
                    mb: 3
                }
            })]
        })]
    }), r.jsx(F, {
        variant: "small",
        color: "textSecondary",
        children: `\u0420\u0430\u0437\u043C\u0435\u0440 \u0444\u0430\u0439\u043B\u0430 \u0434\u043E ${u}`
    })]
})
  , pp = ({dataTest: e="dropzone_area_view", dropzoneProps: t, accept: a, handleDropAccepted: n, handleDropRejected: o, maxFileSize: i, maxFileSizeFormatted: l, isMultiple: s, filesLimit: c, dropzoneText: d, icon: p, inputProps: u, disableRejectionFeedback: f, dropZoneStyles: m, dropzoneParagraphStyles: b}) => r.jsx(ln, {
    ...t,
    accept: a,
    onDropAccepted: n,
    onDropRejected: o,
    maxSize: i,
    multiple: s,
    maxFiles: c,
    children: (y) => r.jsx(cp, {
        ...y,
        dataTest: e,
        dropzoneText: d,
        icon: p,
        inputProps: u,
        disableRejectionFeedback: f,
        dropZoneStyles: m,
        dropzoneParagraphStyles: b,
        maxFileSizeFormatted: l
    })
})
  , dp = (e) => {
    const t = lr(e, "\u0444\u0430\u0439\u043B", "\u0444\u0430\u0439\u043B\u0430", "\u0444\u0430\u0439\u043B\u043E\u0432");
    return `\u0412\u044B \u0443\u0436\u0435 \u0437\u0430\u0433\u0440\u0443\u0437\u0438\u043B\u0438 \u043C\u0430\u043A\u0441\u0438\u043C\u0430\u043B\u044C\u043D\u043E \u0432\u043E\u0437\u043C\u043E\u0436\u043D\u043E\u0435 \u043A\u043E\u043B\u0438\u0447\u0435\u0441\u0442\u0432\u043E \u0444\u0430\u0439\u043B\u043E\u0432. \u0420\u0430\u0437\u0440\u0435\u0448\u0435\u043D\u043E \u0442\u043E\u043B\u044C\u043A\u043E ${e} ${t}`;
}
  , up = (e, t, a) => {
    let n = `\u0424\u0430\u0439\u043B ${e.file.name} \u0431\u044B\u043B \u043E\u0442\u043A\u043B\u043E\u043D\u0435\u043D.`;
    return t.includes(e.file.type) || (n += " \u0414\u0430\u043D\u043D\u044B\u0439 \u0442\u0438\u043F \u0444\u0430\u0439\u043B\u0430 \u043D\u0435 \u043F\u043E\u0434\u0434\u0435\u0440\u0436\u0438\u0432\u0430\u0435\u0442\u0441\u044F."),
    e.file.size > a && (n += ` \u0424\u0430\u0439\u043B \u0441\u043B\u0438\u0448\u043A\u043E\u043C \u0431\u043E\u043B\u044C\u0448\u043E\u0439. \u041C\u0430\u043A\u0441\u0438\u043C\u0430\u043B\u044C\u043D\u044B\u0439 \u0440\u0430\u0437\u043C\u0435\u0440 \u0444\u0430\u0439\u043B\u0430 - ${nc(a)}.`),
    n;
}
  , mp = ({dataTest: e="dropzone_area_base", accept: t={}, filesLimit: a=1, icon: n, fileObjects: o=[], maxFileSize: i=3e7, dropzoneText: l="\u041F\u0435\u0440\u0435\u0442\u0430\u0449\u0438\u0442\u0435 \u0444\u0430\u0439\u043B\u044B \u0441\u044E\u0434\u0430 \u0438\u043B\u0438 \u043D\u0430\u0436\u043C\u0438\u0442\u0435, \u0447\u0442\u043E\u0431\u044B \u0432\u044B\u0431\u0440\u0430\u0442\u044C \u0444\u0430\u0439\u043B\u044B", disableRejectionFeedback: s=false, getFileLimitExceedMessage: c=dp, getDropRejectMessage: d=up, onAdd: p, onDrop: u, onDropRejected: f, inputProps: m, dropzoneProps: b, dropZoneStyles: y=[], dropzoneParagraphStyles: w=[]}) => {
    const E = a > 1
      , A = rc(t)
      , C = va()
      , S = g.useMemo( () => C(i).format({
        output: "byte",
        base: "binary",
        mantissa: 2,
        spaceSeparated: true
    }), [i, C])
      , T = async (_, U) => {
        if (a > 1 && o.length + _.length > a) {
            ro.error(c(a));
            return;
        }
        u == null ? void 0 : u(_, U),
        p == null ? void 0 : p(_);
    }
      , P = (_, U) => {
        let L = "";
        if (o.length + _.length > a)
            L = c(a);
        else
            for (const $ of _)
                L = d($, A, i);
        f == null ? void 0 : f(_, U),
        ro.error(L);
    }
    ;
    return r.jsx(pp, {
        dataTest: e,
        dropzoneProps: b,
        accept: t,
        handleDropAccepted: T,
        handleDropRejected: P,
        maxFileSize: i,
        maxFileSizeFormatted: S,
        isMultiple: E,
        filesLimit: a,
        dropzoneText: l,
        icon: n,
        inputProps: m,
        disableRejectionFeedback: s,
        dropZoneStyles: y,
        dropzoneParagraphStyles: w
    });
}
  , fp = ({dataTest: e="dropzone", accept: t, filesLimit: a=1, icon: n, maxFileSize: o, dropzoneText: i, disableRejectionFeedback: l, getFileAddedMessage: s, getFileLimitExceedMessage: c, getDropRejectMessage: d, onDrop: p, onDropRejected: u, inputProps: f, dropzoneProps: m, dropZoneStyles: b, dropzoneParagraphStyles: y}) => {
    const [w,E] = g.useState([])
      , A = async (C) => {
        const S = C.map( (T) => ({
            name: T.name,
            type: T.type,
            size: T.size,
            path: T.path,
            lastModified: T.lastModified,
            webkitRelativePath: T.webkitRelativePath
        }));
        E([...w, ...S]);
    }
    ;
    return r.jsx(mp, {
        dataTest: e,
        accept: t,
        filesLimit: a,
        icon: n,
        fileObjects: w,
        maxFileSize: o,
        dropzoneText: i,
        disableRejectionFeedback: l,
        getFileAddedMessage: s,
        getFileLimitExceedMessage: c,
        getDropRejectMessage: d,
        onAdd: A,
        onDrop: p,
        onDropRejected: u,
        inputProps: f,
        dropzoneProps: m,
        dropZoneStyles: b,
        dropzoneParagraphStyles: y
    });
}
  , Ii = () => br(cr)
  , vp = (e) => ta(sr(e))
  , gp = ({onDownload: e, onDelete: t, isDisabled: a, isDeletable: n}) => r.jsxs(r.Fragment, {
    children: [r.jsx(rp, {
        sx: {
            ml: "auto",
            fontSize: "2rem"
        },
        onClick: e,
        color: "primary",
        isDisabled: a
    }), n && r.jsx(Jn, {
        variant: "thin",
        onClick: t,
        sx: {
            fontSize: "2rem",
            ml: 2
        },
        isDisabled: a
    })]
})
  , xp = ({filename: e, type: t, size: a, uploader: n, uploadDate: o}) => r.jsxs(K, {
    display: "flex",
    flexDirection: "column",
    children: [r.jsx(F, {
        fontSize: "1rem",
        sx: {
            m: 0,
            p: 0
        },
        children: e
    }), r.jsx(F, {
        fontSize: "0.8rem",
        variant: "small",
        color: "textSecondary",
        sx: {
            m: 0,
            p: 0
        },
        children: t
    }), r.jsx(F, {
        color: "textSecondary",
        sx: {
            mt: "auto"
        },
        lineHeight: 1,
        children: a
    }), r.jsx(F, {
        fontSize: "0.8rem",
        variant: "small",
        color: "textSecondary",
        children: `\u0417\u0430\u0433\u0440\u0443\u0436\u0435\u043D\u043E: ${n} - ${o}`
    })]
})
  , hp = {
    display: "flex",
    minHeight: 75,
    width: "100%",
    py: 3
}
  , bp = ({attachment: e, size: t="", attachmentTypeDisplayValue: a, mimeType: n="", createdAt: o, handleDownload: i, onDelete: l, isDeletable: s, isDisabled: c}) => r.jsx(K, {
    display: "flex",
    width: "100%",
    flexDirection: "column",
    children: r.jsx(Ia, {
        variant: "outlined",
        sx: {
            width: "inherit",
            mb: 2
        },
        children: r.jsxs(_a, {
            sx: hp,
            children: [r.jsx(ji, {
                fileType: n,
                sx: {
                    mr: 2
                },
                styles: {
                    fontSize: "4.5rem"
                }
            }), r.jsxs(K, {
                display: "flex",
                justifyItems: "space-between",
                width: "100%",
                alignItems: "center",
                children: [r.jsx(xp, {
                    filename: (e == null ? void 0 : e.original_filename) || "",
                    type: a,
                    size: t,
                    uploader: (e == null ? void 0 : e.user.full_name) || "",
                    uploadDate: o
                }), r.jsx(gp, {
                    onDownload: i,
                    onDelete: l,
                    isDisabled: c,
                    isDeletable: s
                })]
            })]
        })
    })
})
  , yp = ({id: e, onDelete: t, isDeletable: a, isDisabled: n}) => {
    const {data: o} = Ii()
      , {data: i} = vp(e)
      , l = va()
      , s = g.useMemo( () => {
        if (i) {
            const {file_type: m} = o
              , b = z.find(m, ["value", i.file_type]);
            if (b)
                return b.display_name;
        }
        return "";
    }
    , [i, o])
      , c = g.useMemo( () => {
        if (i) {
            const m = pr.lookup(i.original_filename);
            if (m)
                return m;
        }
        return "";
    }
    , [i])
      , d = g.useMemo( () => i ? l(i.size).format({
        output: "byte",
        base: "binary",
        mantissa: 2,
        spaceSeparated: true
    }) : "", [i, l])
      , p = g.useCallback( () => {
        i && sp(i.file, i.original_filename);
    }
    , [i])
      , u = g.useMemo( () => i ? Ot.fromISO(i.created_at).setLocale("ru").toLocaleString() : Ot.now().setLocale("ru").toLocaleString(), [i])
      , f = g.useCallback( () => {
        t == null ? void 0 : t(e);
    }
    , [t, e]);
    return r.jsx(bp, {
        attachment: i,
        attachmentTypeDisplayValue: s,
        mimeType: c,
        createdAt: u,
        handleDownload: p,
        onDelete: f,
        size: d,
        isDeletable: a,
        isDisabled: n
    });
}
  , wp = (e="") => {
    const t = "FwsAZAvhpk8eQHXG5kQwA7RyDgkrA82Z9Veq6DxO";
    return ta({
        queryKey: nn.externalServices.microlink(e).queryKey,
        queryFn: () => {
            const a = Xo(e, {
                forceHttps: true,
                stripHash: true,
                stripTextFragment: true,
                removeQueryParameters: true,
                removeTrailingSlash: true,
                removeSingleSlash: true,
                removeExplicitPort: true
            });
            return dr(a.href, {
                apiKey: t,
                ttl: "max",
                screenshot: true
            });
        }
        ,
        enabled: !!e,
        gcTime: Number.POSITIVE_INFINITY,
        staleTime: Number.POSITIVE_INFINITY
    });
}
  , eo = () => ta(mr)
  , Cp = (e) => ta(ur(e))
  , Sp = (e, t) => {
    var _a2;
    return !(e == null ? void 0 : e.url_type) || !(t == null ? void 0 : t.url_type) ? "" : ((_a2 = z.find(t.url_type, ["value", e.url_type])) == null ? void 0 : _a2.display_name) ?? "";
}
  , Ep = {
    forceHttps: true,
    stripHash: true,
    stripTextFragment: true,
    removeQueryParameters: true,
    removeTrailingSlash: true,
    removeSingleSlash: true,
    removeExplicitPort: true
}
  , Ap = g.createContext({})
  , _i = true;
function kp({baseColor: e, highlightColor: t, width: a, height: n, borderRadius: o, circle: i, direction: l, duration: s, enableAnimation: c=_i, customHighlightBackground: d}) {
    const p = {};
    return l === "rtl" && (p["--animation-direction"] = "reverse"),
    typeof s == "number" && (p["--animation-duration"] = `${s}s`),
    c || (p["--pseudo-element-display"] = "none"),
    (typeof a == "string" || typeof a == "number") && (p.width = a),
    (typeof n == "string" || typeof n == "number") && (p.height = n),
    (typeof o == "string" || typeof o == "number") && (p.borderRadius = o),
    i && (p.borderRadius = "50%"),
    typeof e < "u" && (p["--base-color"] = e),
    typeof t < "u" && (p["--highlight-color"] = t),
    typeof d == "string" && (p["--custom-highlight-background"] = d),
    p;
}
function On({count: e=1, wrapper: t, className: a, containerClassName: n, containerTestId: o, circle: i=false, style: l, ...s}) {
    var c, d, p;
    const u = g.useContext(Ap)
      , f = {
        ...s
    };
    for (const [C,S] of Object.entries(s))
        typeof S > "u" && delete f[C];
    const m = {
        ...u,
        ...f,
        circle: i
    }
      , b = {
        ...l,
        ...kp(m)
    };
    let y = "react-loading-skeleton";
    a && (y += ` ${a}`);
    const w = (c = m.inline) !== null && c !== void 0 ? c : false
      , E = []
      , A = Math.ceil(e);
    for (let C = 0; C < A; C++) {
        let S = b;
        if (A > e && C === A - 1) {
            const P = (d = S.width) !== null && d !== void 0 ? d : "100%"
              , _ = e % 1
              , U = typeof P == "number" ? P * _ : `calc(${P} * ${_})`;
            S = {
                ...S,
                width: U
            };
        }
        const T = g.createElement("span", {
            className: y,
            style: S,
            key: C
        }, "\u200C");
        w ? E.push(T) : E.push(g.createElement(g.Fragment, {
            key: C
        }, T, g.createElement("br", null)));
    }
    return g.createElement("span", {
        className: n,
        "data-testid": o,
        "aria-live": "polite",
        "aria-busy": (p = m.enableAnimation) !== null && p !== void 0 ? p : _i
    }, t ? E.map( (C, S) => g.createElement(t, {
        key: S
    }, C)) : E);
}
const Tp = ({isPending: e, title: t="", description: a="", urlNormalised: n="", linkTypeDisplayValue: o=""}) => r.jsxs(_a, {
    sx: {
        display: "flex",
        minHeight: 75,
        width: "100%",
        flexDirection: "column"
    },
    children: [r.jsx(F, {
        variant: "strong",
        sx: {
            fontSize: "0.75rem",
            mb: 1
        },
        children: e ? r.jsx(On, {
            containerTestId: "skeleton_title"
        }) : t || "\u041D\u0435\u0442 \u0437\u0430\u0433\u043E\u043B\u043E\u0432\u043A\u0430"
    }), r.jsx(F, {
        sx: {
            fontSize: "0.75rem"
        },
        children: e ? r.jsx(On, {
            count: 4,
            containerTestId: "skeleton_description"
        }) : a || "\u041D\u0435\u0442 \u043E\u043F\u0438\u0441\u0430\u043D\u0438\u044F"
    }), r.jsx(K, {
        sx: {
            mt: 2
        },
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        children: r.jsxs(K, {
            children: [r.jsx(F, {
                fontSize: "inherit",
                sx: {
                    mr: 1
                },
                color: "textSecondary",
                children: n
            }), r.jsx(F, {
                component: "span",
                variant: "small",
                color: "textSecondary",
                fontSize: "inherit",
                children: `(\u0442\u0438\u043F: ${o})`
            })]
        })
    })]
})
  , Pp = () => r.jsx(On, {
    containerTestId: "skeleton_media",
    height: 150
})
  , Op = ({isPending: e, screenshotUrl: t}) => r.jsx(fr, {
    component: e ? Pp : "img",
    image: t,
    sx: {
        height: "auto",
        minHeight: 150
    }
})
  , jp = ({isDisabled: e, onDelete: t}) => r.jsx(Mn, {
    "data-testid": "delete_button",
    sx: {
        position: "absolute",
        right: 10,
        bottom: 13
    },
    onClick: t,
    disabled: e,
    name: "delete",
    type: "button",
    "aria-label": "delete",
    children: r.jsx(Jn, {
        isDisabled: e
    })
})
  , Ip = {
    textDecoration: "none",
    color: "unset",
    cursor: "unset"
}
  , _p = ({url: e="", isDisabled: t, children: a}) => r.jsx(Dr, {
    to: e,
    target: "_blank",
    rel: "noreferrer",
    style: Ip,
    children: r.jsx(js, {
        disabled: t,
        children: a
    })
})
  , Rp = ({url: e="", isPending: t, isDisabled: a, isDeletable: n, screenshotUrl: o="", title: i="", description: l="", urlNormalised: s="", linkTypeDisplayValue: c="", onDelete: d}) => r.jsx(K, {
    display: "flex",
    flexDirection: "column",
    children: r.jsxs(Ia, {
        variant: "outlined",
        sx: {
            position: "relative",
            width: "inherit",
            mb: 2
        },
        children: [r.jsxs(_p, {
            url: e,
            isDisabled: a,
            children: [r.jsx(Op, {
                isPending: t,
                screenshotUrl: o
            }), r.jsx(Tp, {
                isPending: t,
                title: i,
                description: l,
                urlNormalised: s,
                linkTypeDisplayValue: c
            })]
        }), n && r.jsx(jp, {
            isDisabled: a,
            onDelete: d
        })]
    })
})
  , Dp = ({id: e="", isDeletable: t, isDisabled: a, onDelete: n}) => {
    var _a2;
    const {data: o} = eo()
      , {data: i} = Cp(e)
      , {data: l, isPending: s} = wp(i == null ? void 0 : i.url)
      , c = g.useMemo( () => Sp(i, o), [i, o])
      , d = g.useMemo( () => (i == null ? void 0 : i.url) ? Xo(i.url, Ep).resource : "", [i])
      , p = g.useCallback( (u) => {
        u.preventDefault(),
        u.stopPropagation(),
        n == null ? void 0 : n(e);
    }
    , [e, n]);
    return r.jsx(Rp, {
        url: i == null ? void 0 : i.url,
        isPending: s,
        isDisabled: a,
        isDeletable: t,
        screenshotUrl: (_a2 = l == null ? void 0 : l.data.screenshot) == null ? void 0 : _a2.url,
        title: l == null ? void 0 : l.data.title,
        description: l == null ? void 0 : l.data.description,
        urlNormalised: d,
        linkTypeDisplayValue: c,
        onDelete: p
    });
}
  , Fp = ({children: e}) => r.jsx(F, {
    "data-testid": "page_title",
    variant: "display1",
    component: "h1",
    children: e
})
  , oa = ({dataTest: e="sub_title", title: t="Title", supportTitle: a, supportContent: n, variant: o="display6"}) => r.jsxs(K, {
    "data-testid": e,
    display: "flex",
    sx: {
        mb: 1,
        mt: 5
    },
    children: [r.jsx(F, {
        variant: o,
        children: t
    }), a && n && r.jsx(op, {
        title: a,
        content: n
    })]
})
  , Lp = "text_input"
  , Mp = ({isSubmitting: e=false, disabled: t, fieldName: a=Lp, placeholder: n, label: o, sx: i, onEditSave: l, isEditable: s, defaultHelperText: c}) => {
    const d = g.useRef(null)
      , [p,u,f] = Ft(a)
      , m = Ve()
      , b = g.useCallback( () => {
        f.setValue(u.initialValue);
    }
    , [f, u.initialValue])
      , y = g.useCallback( (A) => {
        f.setValue(A.target.value);
    }
    , [f])
      , w = g.useCallback( (A) => {
        A.preventDefault(),
        m.handleBlur(A);
    }
    , [m])
      , E = a.replaceAll(".", "_");
    return Bn( () => {
        f.setValue(u.initialValue).catch(console.error);
    }
    ),
    r.jsx(aa, {
        name: E,
        children: () => r.jsx(Nn, {
            ref: d,
            isEditable: s,
            onEditSave: l,
            onEditCancel: b,
            onChange: y,
            onBlur: w,
            name: p.name,
            dataTest: E,
            value: p.value,
            label: o,
            id: E,
            type: "text",
            disabled: e || t,
            placeholder: n,
            sx: i,
            error: u.touched && !!u.error,
            helperText: u.touched && u.error || c
        })
    });
}
  , $p = "file"
  , Bp = ({fieldName: e=$p, fileType: t=""}) => r.jsx(aa, {
    name: e,
    children: () => r.jsx(ji, {
        dataTest: `file_type_icon_${e}`,
        fileType: t,
        styles: {
            cursor: "none",
            pointerEvents: "none",
            fontSize: "4rem"
        }
    })
})
  , Np = ({dataTest: e="input_counter", characterMax: t, character: a, sx: n}) => {
    const o = g.useMemo( () => {
        const i = `${a}`
          , l = t && `/${t}`;
        return t ? `${i}${l}` : `${i}`;
    }
    , [a, t]);
    return r.jsx(F, {
        variant: "small",
        "data-testid": e,
        color: t && a <= t ? "textPrimary" : "error",
        sx: {
            position: "absolute",
            right: 0,
            bottom: 2,
            ...n
        },
        children: o
    });
}
  , zp = "text_area_with_counter"
  , Vp = ({maxSize: e, isSubmitting: t=false, rows: a=5, placeholder: n, label: o, disabled: i, fieldName: l=zp, sx: s, onEditSave: c, isEditable: d}) => {
    const p = g.useRef(null)
      , [u,f,m] = Ft(l)
      , b = Ve()
      , y = g.useCallback( (A) => {
        A.preventDefault(),
        b.handleBlur(A);
    }
    , [b])
      , w = g.useCallback( () => {
        m.setValue(f.initialValue);
    }
    , [m, f.initialValue])
      , E = g.useCallback( (A) => {
        m.setValue(A.target.value);
    }
    , [m]);
    return r.jsx(aa, {
        name: l,
        id: l,
        children: () => {
            var _a2;
            const A = ((_a2 = u.value) == null ? void 0 : _a2.length) || 0;
            return r.jsxs(K, {
                "data-testid": `${l}_field_wrapper`,
                display: "flex",
                flexDirection: "column",
                sx: {
                    position: "relative",
                    ...s
                },
                children: [r.jsx(Nn, {
                    dataTest: `${l}_field`,
                    ref: p,
                    isEditable: d,
                    onEditSave: c,
                    onEditCancel: w,
                    onBlur: y,
                    onChange: E,
                    name: u.name,
                    multiline: true,
                    "data-testid": u.name,
                    value: u.value,
                    label: o,
                    id: u.name,
                    placeholder: n,
                    disabled: t || i,
                    minRows: a,
                    helperText: ri(f.error) || void 0,
                    sx: {
                        width: "100%",
                        "& .MuiInputBase-root": {
                            alignItems: "baseline"
                        }
                    },
                    error: f.touched && !!f.error
                }), r.jsx(Np, {
                    characterMax: e,
                    character: A
                })]
            });
        }
    });
}
  , Hp = "predefined_choice"
  , Ri = ({dataTest: e="predefined_choice_select_field", isSubmitting: t=false, disabled: a, fieldName: n=Hp, label: o="Predefined choice select field", choices: i, readOnly: l, variant: s="outlined", sx: c, isEditable: d, onSelect: p, onBlur: u}) => {
    const [f,m,b] = Ft(n)
      , y = Ve()
      , w = n.replaceAll(".", "_")
      , E = g.useCallback( (C) => {
        const {value: S} = C.target
          , T = z.find(i, ["value", S]);
        T && (b.setValue(T),
        p == null ? void 0 : p(T));
    }
    , [i, b, p])
      , A = g.useCallback( (C) => {
        C.preventDefault(),
        u == null ? void 0 : u(C),
        y.handleBlur(C);
    }
    , [y, u]);
    return r.jsx(aa, {
        name: w,
        children: () => {
            var _a2;
            return r.jsxs(Qo, {
                variant: s,
                "data-testid": e,
                sx: {
                    width: "35%",
                    ...c
                },
                children: [r.jsx(vr, {
                    id: e,
                    children: o
                }), r.jsx(gr, {
                    labelId: e,
                    value: ((_a2 = f.value) == null ? void 0 : _a2.value) || "",
                    name: f.name,
                    "data-testid": `${w}_select`,
                    label: o,
                    disabled: t || a,
                    readOnly: l,
                    variant: s,
                    onChange: E,
                    onBlur: A,
                    error: m.touched && !!m.error,
                    sx: {
                        cursor: d || t || a ? "not-allowed" : "pointer",
                        pointerEvents: d || t || a ? "none" : "auto"
                    },
                    children: i.map( (C) => r.jsx(jr, {
                        value: C.value,
                        children: C.display_name
                    }, C.value))
                }), r.jsx(Zo, {
                    error: m.touched && !!m.error,
                    children: m.touched && ri(m.error) || " "
                })]
            });
        }
    });
}
  , Up = "numeral"
  , qp = ({isSubmitting: e=false, disabled: t, label: a, placeholder: n, fieldName: o=Up, prefix: i, suffix: l, displayFormat: s, styles: c, sx: d, isEditable: p, onEditSave: u}) => {
    const f = g.useRef(null)
      , [m,b,y] = Ft(o)
      , [w,E] = g.useState("")
      , A = Ve()
      , C = va()
      , S = g.useCallback( () => {
        y.setValue(b.initialValue),
        E(z.isFinite(z.toNumber(b.initialValue)) ? C(b.initialValue).format(s) : "");
    }
    , [s, y, b.initialValue, C])
      , T = g.useCallback( (L) => {
        const {value: $} = L.target;
        $ && E(C($).format({
            thousandSeparated: false
        }));
    }
    , [C])
      , P = g.useCallback(async (L) => {
        L.preventDefault(),
        E(z.isFinite(m.value) ? C(m.value).format(s) : ""),
        A.handleBlur(L);
    }
    , [s, m.value, A, C])
      , _ = g.useCallback( (L) => {
        let q = L.target.value;
        const M = /^\d+([.,]\d{0,4})?$/;
        q = q.replace(",", "."),
        M.test(q) && (y.setValue(z.toNumber(q)),
        E(q)),
        q === "" && (y.setValue(void 0),
        E(q));
    }
    , [y]);
    Bn( () => {
        z.isFinite(b.initialValue) ? E(C(b.initialValue).format(s)) : E("");
    }
    );
    const U = o.replaceAll(".", "_");
    return r.jsx(aa, {
        name: o,
        children: () => r.jsx(Nn, {
            ref: f,
            isEditable: p,
            onEditSave: u,
            onEditCancel: S,
            onBlur: P,
            onChange: _,
            onDoubleClick: T,
            name: m.name,
            dataTest: U,
            value: w,
            label: a,
            id: U,
            type: "text",
            placeholder: n,
            style: c,
            sx: d,
            disabled: e || t,
            error: b.touched && !!b.error,
            helperText: b.touched && b.error || void 0,
            slotProps: {
                input: {
                    startAdornment: i && r.jsx(lo, {
                        position: "start",
                        children: i
                    }),
                    endAdornment: l && r.jsx(lo, {
                        position: "end",
                        children: l
                    })
                }
            }
        })
    });
}
  , Wp = "financial_data"
  , Yp = ({isSubmitting: e=false, disabled: t, label: a, placeholder: n, fieldName: o=Wp, prefix: i, currency: l, onEditSave: s, isEditable: c, sx: d}) => r.jsx(qp, {
    placeholder: n,
    isSubmitting: e,
    disabled: t,
    label: a,
    fieldName: o,
    prefix: i,
    suffix: l,
    displayFormat: {
        thousandSeparated: true,
        mantissa: 2
    },
    onEditSave: s,
    isEditable: c,
    sx: d
})
  , Gp = "text/plain"
  , Kp = "us-ascii"
  , vn = (e, t) => t.some( (a) => a instanceof RegExp ? a.test(e) : a === e)
  , Xp = /* @__PURE__ */
new Set(["https:", "http:", "file:"])
  , Qp = (e) => {
    try {
        const {protocol: t} = new URL(e);
        return t.endsWith(":") && !t.includes(".") && !Xp.has(t);
    } catch {
        return false;
    }
}
  , Zp = (e, {stripHash: t}) => {
    var _a2;
    const a = /^data:(?<type>[^,]*?),(?<data>[^#]*?)(?:#(?<hash>.*))?$/.exec(e);
    if (!a)
        throw new Error(`Invalid URL: ${e}`);
    let {type: n, data: o, hash: i} = a.groups;
    const l = n.split(";");
    i = t ? "" : i;
    let s = false;
    l[l.length - 1] === "base64" && (l.pop(),
    s = true);
    const c = ((_a2 = l.shift()) == null ? void 0 : _a2.toLowerCase()) ?? ""
      , p = [...l.map( (u) => {
        let[f,m=""] = u.split("=").map( (b) => b.trim());
        return f === "charset" && (m = m.toLowerCase(),
        m === Kp) ? "" : `${f}${m ? `=${m}` : ""}`;
    }
    ).filter(Boolean)];
    return s && p.push("base64"),
    (p.length > 0 || c && c !== Gp) && p.unshift(c),
    `data:${p.join(";")},${s ? o.trim() : o}${i ? `#${i}` : ""}`;
}
;
function Jp(e, t) {
    if (t = {
        defaultProtocol: "http",
        normalizeProtocol: true,
        forceHttp: false,
        forceHttps: false,
        stripAuthentication: true,
        stripHash: false,
        stripTextFragment: true,
        stripWWW: true,
        removeQueryParameters: [/^utm_\w+/i],
        removeTrailingSlash: true,
        removeSingleSlash: true,
        removeDirectoryIndex: false,
        removeExplicitPort: false,
        sortQueryParameters: true,
        ...t
    },
    typeof t.defaultProtocol == "string" && !t.defaultProtocol.endsWith(":") && (t.defaultProtocol = `${t.defaultProtocol}:`),
    e = e.trim(),
    /^data:/i.test(e))
        return Zp(e, t);
    if (Qp(e))
        return e;
    const a = e.startsWith("//");
    !a && /^\.*\//.test(e) || (e = e.replace(/^(?!(?:\w+:)?\/\/)|^\/\//, t.defaultProtocol));
    const o = new URL(e);
    if (t.forceHttp && t.forceHttps)
        throw new Error("The `forceHttp` and `forceHttps` options cannot be used together");
    if (t.forceHttp && o.protocol === "https:" && (o.protocol = "http:"),
    t.forceHttps && o.protocol === "http:" && (o.protocol = "https:"),
    t.stripAuthentication && (o.username = "",
    o.password = ""),
    t.stripHash ? o.hash = "" : t.stripTextFragment && (o.hash = o.hash.replace(/#?:~:text.*?$/i, "")),
    o.pathname) {
        const l = /\b[a-z][a-z\d+\-.]{1,50}:\/\//g;
        let s = 0
          , c = "";
        for (; ; ) {
            const p = l.exec(o.pathname);
            if (!p)
                break;
            const u = p[0]
              , f = p.index
              , m = o.pathname.slice(s, f);
            c += m.replace(/\/{2,}/g, "/"),
            c += u,
            s = f + u.length;
        }
        const d = o.pathname.slice(s, o.pathname.length);
        c += d.replace(/\/{2,}/g, "/"),
        o.pathname = c;
    }
    if (o.pathname)
        try {
            o.pathname = decodeURI(o.pathname);
        } catch {}
    if (t.removeDirectoryIndex === true && (t.removeDirectoryIndex = [/^index\.[a-z]+$/]),
    Array.isArray(t.removeDirectoryIndex) && t.removeDirectoryIndex.length > 0) {
        let l = o.pathname.split("/");
        const s = l[l.length - 1];
        vn(s, t.removeDirectoryIndex) && (l = l.slice(0, -1),
        o.pathname = l.slice(1).join("/") + "/");
    }
    if (o.hostname && (o.hostname = o.hostname.replace(/\.$/, ""),
    t.stripWWW && /^www\.(?!www\.)[a-z\-\d]{1,63}\.[a-z.\-\d]{2,63}$/.test(o.hostname) && (o.hostname = o.hostname.replace(/^www\./, ""))),
    Array.isArray(t.removeQueryParameters))
        for (const l of [...o.searchParams.keys()])
            vn(l, t.removeQueryParameters) && o.searchParams.delete(l);
    if (!Array.isArray(t.keepQueryParameters) && t.removeQueryParameters === true && (o.search = ""),
    Array.isArray(t.keepQueryParameters) && t.keepQueryParameters.length > 0)
        for (const l of [...o.searchParams.keys()])
            vn(l, t.keepQueryParameters) || o.searchParams.delete(l);
    if (t.sortQueryParameters) {
        o.searchParams.sort();
        try {
            o.search = decodeURIComponent(o.search);
        } catch {}
    }
    t.removeTrailingSlash && (o.pathname = o.pathname.replace(/\/$/, "")),
    t.removeExplicitPort && o.port && (o.port = "");
    const i = e;
    return e = o.toString(),
    !t.removeSingleSlash && o.pathname === "/" && !i.endsWith("/") && o.hash === "" && (e = e.replace(/\/$/, "")),
    (t.removeTrailingSlash || o.pathname === "/") && o.hash === "" && t.removeSingleSlash && (e = e.replace(/\/$/, "")),
    a && !t.normalizeProtocol && (e = e.replace(/^http:\/\//, "//")),
    t.stripProtocol && (e = e.replace(/^(?:https?:)?\/\//, "")),
    e;
}
const ed = "url"
  , td = ({isSubmitting: e=false, disabled: t, label: a, placeholder: n, fieldName: o=ed, sx: i}) => {
    const [l,s,c] = Ft(o)
      , d = Ve()
      , p = g.useCallback( (m) => {
        c.setValue(m.target.value);
    }
    , [c])
      , u = g.useCallback( (m) => {
        if (l.value) {
            const b = Jp(l.value, {
                forceHttps: true,
                stripHash: true,
                stripTextFragment: true,
                removeQueryParameters: true,
                removeTrailingSlash: true,
                removeSingleSlash: true,
                removeExplicitPort: true,
                stripAuthentication: false
            });
            c.setValue(b);
        }
        d.handleBlur(m);
    }
    , [l.value, d, c])
      , f = o.replaceAll(".", "_");
    return r.jsx(aa, {
        name: f,
        children: () => r.jsx(Jo, {
            name: l.name,
            dataTest: f,
            onBlur: u,
            onChange: p,
            value: l.value,
            label: a,
            id: f,
            type: "url",
            disabled: e || t,
            placeholder: n,
            sx: i,
            error: s.touched && !!s.error,
            helperText: s.touched && s.error || " "
        })
    });
}
  , ad = ({label: e="", isOpen: t}) => r.jsxs(K, {
    display: "flex",
    alignItems: "center",
    width: "-webkit-fill-available",
    children: [r.jsx(F, {
        variant: "small",
        sx: {
            width: "100%"
        },
        children: e
    }), r.jsx(Wt, {
        sx: {
            ml: "auto"
        },
        icon: t ? je.caretUpThin : je.caretDownThin
    })]
})
  , nd = "predefined_chip_choice"
  , od = ({dataTest: e="predefined_choice_chip_field", isSubmitting: t=false, disabled: a, fieldName: n=nd, choices: o, emptyLabel: i, emptyColor: l, emptyVariant: s, onChange: c}) => {
    const [d,p] = g.useState()
      , [u,,f] = Ft(n)
      , m = n.replaceAll(".", "_")
      , b = (E) => {
        p(E.currentTarget);
    }
      , y = () => {
        p(void 0);
    }
      , w = g.useCallback( (E) => () => {
        f.setValue(E),
        c == null ? void 0 : c(n, E),
        p(void 0);
    }
    , [n, f, c]);
    return r.jsx(aa, {
        name: m,
        children: () => r.jsxs(r.Fragment, {
            children: [r.jsx(Qn, {
                "data-testid": e,
                label: r.jsx(ad, {
                    label: u.value.display_name || i,
                    isOpen: !!d
                }),
                size: "small",
                onClick: b,
                color: u.value.display_name ? "default" : l || "error",
                variant: u.value.display_name ? "outlined" : s || "filled",
                disabled: t || a,
                sx: {
                    cursor: t || a ? "not-allowed" : "pointer",
                    pointerEvents: t || a ? "none" : "auto"
                },
                clickable: true
            }), r.jsx(Ir, {
                id: `${m}_menu`,
                anchorEl: d,
                onClose: y,
                children: o.map( (E) => r.jsx(_r, {
                    label: E.display_name,
                    onClick: w(E)
                }, Ta()))
            })]
        })
    });
}
  , Bo = "search_autocomplete_field"
  , id = ({label: e, placeholder: t, sx: a, isEditable: n=true, fieldName: o=Bo, type: i="text", value: l, onInputChange: s, options: c=[], isLoading: d, loadingText: p, noOptionsText: u, renderOption: f, getOptionLabel: m, filterOptions: b, getOptionKey: y, onSelect: w, onClear: E}) => {
    const A = Ft(o)[2]
      , C = g.useCallback( (S, T, P) => {
        P === "selectOption" && (A.setValue(T),
        w == null ? void 0 : w(S, T, P)),
        P === "clear" && (A.setValue(void 0),
        E == null ? void 0 : E(S, T, P));
    }
    , [A, E, w]);
    return r.jsx(As, {
        id: Bo,
        filterOptions: b,
        options: c,
        onInputChange: s,
        inputValue: l,
        onChange: C,
        readOnly: !n,
        renderOption: f,
        getOptionLabel: m,
        getOptionKey: y,
        renderInput: (S) => r.jsx(Jo, {
            ...S,
            label: e,
            placeholder: t,
            type: i
        }),
        sx: a,
        loading: d,
        loadingText: p,
        noOptionsText: u,
        fullWidth: true
    });
}
  , rd = /^((ftp|http|https):\/\/)?(www.)?(?!.*(ftp|http|https|www.))[\wА-я-]+(\.[\wА-я]+)+((\/)[\w#]+)*(\/\w+\?[\wА-я]+=\w+(&[\wА-я]+=\w+)*)?$/gm
  , N = {
    CASH_BOX_INITIAL_BALANCE: "CASH_BOX_INITIAL_BALANCE",
    CASH_BOX_FINAL_BALANCE: "CASH_BOX_FINAL_BALANCE",
    PROCEEDS_FROM_CUSTOMER_SALES: "PROCEEDS_FROM_CUSTOMER_SALES",
    DEPOSIT_INTEREST_RECEIPTS: "DEPOSIT_INTEREST_RECEIPTS",
    PROCEEDS_OF_GRANTS_SUBSIDIES: "PROCEEDS_OF_GRANTS_SUBSIDIES",
    PROCEEDS_FROM_CREDIT_FACILITIES: "PROCEEDS_FROM_CREDIT_FACILITIES",
    PROCEEDS_FROM_INVESTORS: "PROCEEDS_FROM_INVESTORS",
    OTHER_PROCEEDS: "OTHER_PROCEEDS",
    INFLOWS: "INFLOWS",
    EMPLOYEE_PAYMENTS: "EMPLOYEE_PAYMENTS",
    TAXES_FROM_PAYROLL: "TAXES_FROM_PAYROLL",
    PURCHASE_OF_RESALE_EQUIPMENT: "PURCHASE_OF_RESALE_EQUIPMENT",
    IT_INFRASTRUCTURE_RENT: "IT_INFRASTRUCTURE_RENT",
    REGULAR_PAYMENTS_TO_SUPPLIERS: "REGULAR_PAYMENTS_TO_SUPPLIERS",
    OTHER_PAYMENTS_TO_SUPPLIERS: "OTHER_PAYMENTS_TO_SUPPLIERS",
    INTEREST_PAYMENTS_ON_LOANS: "INTEREST_PAYMENTS_ON_LOANS",
    REPAYMENT_OF_DEBT_BODY: "REPAYMENT_OF_DEBT_BODY",
    TAXES: "TAXES",
    DIVIDEND_PAYOUTS: "DIVIDEND_PAYOUTS",
    OTHER_PAYMENTS: "OTHER_PAYMENTS",
    OUTFLOWS: "OUTFLOWS",
    DEBT_INITIAL_BALANCE: "DEBT_INITIAL_BALANCE",
    DEBT_FINAL_BALANCE: "DEBT_FINAL_BALANCE"
}
  , gt = [{
    key: N.CASH_BOX_INITIAL_BALANCE,
    row: 2
}, {
    key: N.CASH_BOX_FINAL_BALANCE,
    row: 3
}, {
    key: N.PROCEEDS_FROM_CUSTOMER_SALES,
    row: 5
}, {
    key: N.DEPOSIT_INTEREST_RECEIPTS,
    row: 6
}, {
    key: N.PROCEEDS_OF_GRANTS_SUBSIDIES,
    row: 7
}, {
    key: N.PROCEEDS_FROM_CREDIT_FACILITIES,
    row: 8
}, {
    key: N.PROCEEDS_FROM_INVESTORS,
    row: 9
}, {
    key: N.OTHER_PROCEEDS,
    row: 10
}, {
    key: N.INFLOWS,
    row: 11
}, {
    key: N.EMPLOYEE_PAYMENTS,
    row: 13
}, {
    key: N.TAXES_FROM_PAYROLL,
    row: 14
}, {
    key: N.PURCHASE_OF_RESALE_EQUIPMENT,
    row: 15
}, {
    key: N.IT_INFRASTRUCTURE_RENT,
    row: 16
}, {
    key: N.REGULAR_PAYMENTS_TO_SUPPLIERS,
    row: 17
}, {
    key: N.OTHER_PAYMENTS_TO_SUPPLIERS,
    row: 18
}, {
    key: N.INTEREST_PAYMENTS_ON_LOANS,
    row: 19
}, {
    key: N.REPAYMENT_OF_DEBT_BODY,
    row: 20
}, {
    key: N.TAXES,
    row: 21
}, {
    key: N.DIVIDEND_PAYOUTS,
    row: 22
}, {
    key: N.OTHER_PAYMENTS,
    row: 23
}, {
    key: N.OUTFLOWS,
    row: 24
}, {
    key: N.DEBT_INITIAL_BALANCE,
    row: 26
}, {
    key: N.DEBT_FINAL_BALANCE,
    row: 27
}]
  , Di = h.createContext(null);
function ld() {
    const [e,t] = h.useState(null);
    return h.useEffect( () => {
        t(`mui-p-${Math.round(Math.random() * 1e5)}`);
    }
    , []),
    e;
}
function sd(e) {
    const {children: t, value: a} = e
      , n = ld()
      , o = h.useMemo( () => ({
        idPrefix: n,
        value: a
    }), [n, a]);
    return r.jsx(Di.Provider, {
        value: o,
        children: t
    });
}
function Fi() {
    return h.useContext(Di);
}
function Li(e, t) {
    const {idPrefix: a} = e;
    return a === null ? null : `${e.idPrefix}-P-${t}`;
}
function Mi(e, t) {
    const {idPrefix: a} = e;
    return a === null ? null : `${e.idPrefix}-T-${t}`;
}
const cd = h.forwardRef(function(t, a) {
    const {children: n, ...o} = t
      , i = Fi();
    if (i === null)
        throw new TypeError("No TabContext provided");
    const l = h.Children.map(n, (s) => h.isValidElement(s) ? h.cloneElement(s, {
        "aria-controls": Li(i, s.props.value),
        id: Mi(i, s.props.value)
    }) : null);
    return r.jsx(ac, {
        ...o,
        ref: a,
        value: i.value,
        children: l
    });
});
function pd(e) {
    return _t("MuiTabPanel", e);
}
Rt("MuiTabPanel", ["root", "hidden"]);
const dd = (e) => {
    const {classes: t, hidden: a} = e;
    return Dt({
        root: ["root", a && "hidden"]
    }, pd, t);
}
  , ud = xe("div", {
    name: "MuiTabPanel",
    slot: "Root"
})( ({theme: e}) => ({
    padding: e.spacing(3)
}))
  , md = h.forwardRef(function(t, a) {
    const n = Lr({
        props: t,
        name: "MuiTabPanel"
    })
      , {children: o, className: i, value: l, keepMounted: s=false, ...c} = n
      , d = {
        ...n
    }
      , p = dd(d)
      , u = Fi();
    if (u === null)
        throw new TypeError("No TabContext provided");
    const f = Li(u, l)
      , m = Mi(u, l);
    return r.jsx(ud, {
        "aria-labelledby": m,
        className: Se(p.root, i),
        hidden: l !== u.value,
        id: f,
        ref: a,
        role: "tabpanel",
        ownerState: d,
        ...c,
        children: (s || l === u.value) && o
    });
})
  , Lt = g.createContext({
    isEditable: void 0
})
  , fd = ({children: e}) => {
    const {applicationId: t} = Fe({
        from: Ie
    })
      , {data: a} = Nt(t)
      , n = g.useMemo( () => ({
        isEditable: a == null ? void 0 : a.is_editable
    }), [a == null ? void 0 : a.is_editable]);
    return r.jsx(Lt.Provider, {
        value: n,
        children: e
    });
}
  , Wa = (e) => {
    const t = z.fill(Array.from({
        length: 14
    }), "");
    return t[0] = e,
    t;
}
  , vd = (e) => {
    const t = z.find(e, {
        key: N.CASH_BOX_INITIAL_BALANCE
    })
      , a = z.find(gt, {
        key: N.CASH_BOX_INITIAL_BALANCE
    })
      , n = z.find(gt, {
        key: N.CASH_BOX_FINAL_BALANCE
    });
    return ["\u0414\u0435\u043D\u0435\u0436\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043D\u0430\u0447\u0430\u043B\u043E \u043F\u0435\u0440\u0438\u043E\u0434\u0430", (t == null ? void 0 : t.value) || void 0, ...Array.from({
        length: 12 - 1
    }).map( (i, l) => `=${nt.simpleCellAddressToString({
        sheet: 0,
        col: l + 1,
        row: n.row
    }, 0)}`), `=${nt.simpleCellAddressToString({
        sheet: 0,
        col: 1,
        row: a.row
    }, 0)}`];
}
  , gd = () => {
    const e = z.find(gt, {
        key: N.CASH_BOX_INITIAL_BALANCE
    })
      , t = z.find(gt, {
        key: N.INFLOWS
    })
      , a = z.find(gt, {
        key: N.OUTFLOWS
    });
    return ["\u0414\u0435\u043D\u0435\u0436\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043A\u043E\u043D\u0435\u0446 \u043F\u0435\u0440\u0438\u043E\u0434\u0430", ...Array.from({
        length: 13
    }).map( (n, o) => {
        const i = nt.simpleCellAddressToString({
            col: o + 1,
            row: e.row,
            sheet: 0
        }, 0)
          , l = nt.simpleCellAddressToString({
            col: o + 1,
            row: t.row,
            sheet: 0
        }, 0)
          , s = nt.simpleCellAddressToString({
            col: o + 1,
            row: a.row,
            sheet: 0
        }, 0);
        return `=${i}+${l}-${s}`;
    }
    )];
}
  , xd = (e) => {
    const t = z.find(e, {
        key: N.DEBT_INITIAL_BALANCE
    })
      , a = z.find(gt, {
        key: N.DEBT_INITIAL_BALANCE
    })
      , n = z.find(gt, {
        key: N.DEBT_FINAL_BALANCE
    });
    return ["\u0417\u0430\u0435\u043C\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043D\u0430\u0447\u0430\u043B\u043E \u043F\u0435\u0440\u0438\u043E\u0434\u0430", (t == null ? void 0 : t.value) || void 0, ...Array.from({
        length: 12 - 1
    }).map( (i, l) => `=${nt.simpleCellAddressToString({
        sheet: 0,
        col: l + 1,
        row: n.row
    }, 0)}`), `=${nt.simpleCellAddressToString({
        sheet: 0,
        col: 1,
        row: a.row
    }, 0)}`];
}
  , hd = () => {
    const e = z.find(gt, {
        key: N.DEBT_INITIAL_BALANCE
    })
      , t = z.find(gt, {
        key: N.PROCEEDS_FROM_CREDIT_FACILITIES
    })
      , a = z.find(gt, {
        key: N.REPAYMENT_OF_DEBT_BODY
    });
    return ["\u0417\u0430\u0435\u043C\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043A\u043E\u043D\u0435\u0446 \u043F\u0435\u0440\u0438\u043E\u0434\u0430", ...Array.from({
        length: 13
    }).map( (n, o) => {
        const i = nt.simpleCellAddressToString({
            col: o + 1,
            row: e.row,
            sheet: 0
        }, 0)
          , l = nt.simpleCellAddressToString({
            col: o + 1,
            row: t.row,
            sheet: 0
        }, 0)
          , s = nt.simpleCellAddressToString({
            col: o + 1,
            row: a.row,
            sheet: 0
        }, 0);
        return `=${i}+${l}-${s}`;
    }
    )];
}
  , Ue = (e, t, a, n) => {
    const i = z.filter(t, {
        year: a,
        key: n
    })
      , l = z.find(gt, {
        key: n
    });
    return [e, ...Array.from({
        length: 12
    }).map( (s, c) => {
        var _a2;
        return ((_a2 = z.find(i, {
            month: c + 1
        })) == null ? void 0 : _a2.value) ?? void 0;
    }
    ), `=SUM(${nt.simpleCellRangeToString({
        start: {
            sheet: 0,
            col: 1,
            row: l.row
        },
        end: {
            sheet: 0,
            col: 12,
            row: l.row
        }
    }, 0)})`];
}
  , No = (e, t, a) => [e, ...Array.from({
    length: 12
}).map( (o, i) => `=SUM(${nt.simpleCellRangeToString({
    start: {
        sheet: 0,
        col: i + 1,
        row: t
    },
    end: {
        sheet: 0,
        col: i + 1,
        row: a
    }
}, 0)})`), `=SUM(${nt.simpleCellRangeToString({
    start: {
        sheet: 0,
        col: 1,
        row: t
    },
    end: {
        sheet: 0,
        col: 12,
        row: a
    }
}, 0)})`]
  , bd = (e, t) => ({
    row: e,
    col: t,
    type: "text",
    readOnly: true,
    className: "header-cell"
})
  , yd = (e, t) => ({
    row: e,
    col: t,
    type: "text",
    readOnly: true,
    className: "section-header-cell"
})
  , Ma = (e, t) => ({
    row: e,
    col: t,
    type: "text",
    allowEmpty: true,
    className: "first-column-cell"
})
  , wd = {
    pattern: {
        output: "number",
        mantissa: 2,
        thousandSeparated: true
    },
    culture: "ru-RU"
}
  , $i = (e, t, a, n, o, i, l) => {
    Reflect.apply(Rr.renderers.NumericRenderer, void 0, [e, t, a, n, o, i, l]);
    const s = l.numericFormat;
    Qa.setLanguage((s == null ? void 0 : s.culture) || "ru-RU"),
    t.textContent = z.isNumber(i) ? Qa(i).format((s == null ? void 0 : s.pattern) || {
        output: "number",
        mantissa: 2,
        thousandSeparated: true
    }) : "";
}
  , Cd = (e, t, a, n, o, i, l) => {
    $i(e, t, a, n, o, i, l),
    z.isNil(i) && l.readOnly !== true && (t.className = Se(t.className, "make-me-required"));
}
  , xt = (e, t) => ({
    type: "numeric",
    row: e,
    col: t,
    allowEmpty: true,
    className: "numeric-input-cell",
    renderer: $i,
    numericFormat: wd
})
  , zo = (e) => ({
    ...e,
    renderer: Cd
})
  , to = (e) => ({
    ...e,
    className: Se(e.className, "first-input-cell")
})
  , ea = (e) => ({
    ...e,
    readOnly: true,
    className: "numeric-calculated-cell"
})
  , Sd = () => [{
    row: 0,
    col: 0,
    type: "text",
    readOnly: true,
    className: "header-cell first-column-cell"
}, ...Array.from({
    length: 13
}).map( (e, t) => bd(0, t + 1))]
  , Ya = (e) => Array.from({
    length: 14
}).map( (t, a) => yd(e, a))
  , Vo = (e) => [Ma(e, 0), to(xt(e, 1)), ...Array.from({
    length: 11
}).map( (a, n) => ea(xt(e, n + 2))), ea(xt(e, 13))]
  , Ho = (e) => [Ma(e, 0), ...Array.from({
    length: 12
}).map( (a, n) => ea(xt(e, n + 1))), ea(xt(e, 13))]
  , Ye = (e) => [Ma(e, 0), to(xt(e, 1)), ...Array.from({
    length: 11
}).map( (a, n) => xt(e, n + 2)), ea(xt(e, 13))]
  , Ed = (e) => [Ma(e, 0), to(zo(xt(e, 1))), ...Array.from({
    length: 11
}).map( (a, n) => zo(xt(e, n + 2))), ea(xt(e, 13))]
  , Uo = (e) => [Ma(e, 0), ...Array.from({
    length: 13
}).map( (a, n) => ea(xt(e, n + 1)))]
  , Ad = (e, t) => {
    const a = Ot.now()
      , n = a.year
      , o = a.month
      , i = [];
    if (i.push({
        type: "text",
        readOnly: true
    }),
    n > e)
        for (let l = 1; l <= 12; l += 1)
            i.push({
                readOnly: !t,
                readOnlyCellClassName: "read-only-cell"
            });
    else
        for (let l = 1; l <= 12; l += 1)
            i.push({
                readOnly: !t || l >= o,
                readOnlyCellClassName: "read-only-cell"
            });
    return i.push({
        readOnly: true
    }),
    i;
}
  , kd = () => {
    const e = [];
    return e.push(...Sd(), ...Ya(1), ...Vo(2), ...Ho(3), ...Ya(4), ...Ed(5), ...Ye(6), ...Ye(7), ...Ye(8), ...Ye(9), ...Ye(10), ...Uo(11), ...Ya(12), ...Ye(13), ...Ye(14), ...Ye(15), ...Ye(16), ...Ye(17), ...Ye(18), ...Ye(19), ...Ye(20), ...Ye(21), ...Ye(22), ...Ye(23), ...Uo(24), ...Ya(25), ...Vo(26), ...Ho(27)),
    e;
}
  , Td = (e, t=[]) => {
    const a = [`${e} \u0433., \u0440\u0443\u0431.`, "\u044F\u043D\u0432", "\u0444\u0435\u0432", "\u043C\u0430\u0440", "\u0430\u043F\u0440", "\u043C\u0430\u0439", "\u0438\u044E\u043D", "\u0438\u044E\u043B", "\u0430\u0432\u0433", "\u0441\u0435\u043D", "\u043E\u043A\u0442", "\u043D\u043E\u044F", "\u0434\u0435\u043A", "\u0418\u0442\u043E\u0433\u043E"]
      , n = z.filter(t, {
        year: e
    });
    return [a, Wa("\u0414\u0435\u043D\u0435\u0436\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430"), vd(n), gd(), Wa("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F"), Ue("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F \u043E\u0442 \u043A\u043B\u0438\u0435\u043D\u0442\u043E\u0432 \u0441 \u043F\u0440\u043E\u0434\u0430\u0436", n, e, N.PROCEEDS_FROM_CUSTOMER_SALES), Ue("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F \u043F\u0440\u043E\u0446\u0435\u043D\u0442\u043E\u0432 \u043F\u043E \u0434\u0435\u043F\u043E\u0437\u0438\u0442\u0430\u043C", n, e, N.DEPOSIT_INTEREST_RECEIPTS), Ue("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F \u0433\u0440\u0430\u043D\u0442\u043E\u0432, \u0441\u0443\u0431\u0441\u0438\u0434\u0438\u0439 \u0438 \u0442.\u043F.", n, e, N.PROCEEDS_OF_GRANTS_SUBSIDIES), Ue("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u0435 \u043A\u0440\u0435\u0434\u0438\u0442\u043D\u044B\u0445 \u0441\u0440\u0435\u0434\u0441\u0442\u0432", n, e, N.PROCEEDS_FROM_CREDIT_FACILITIES), Ue("\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432 \u043E\u0442 \u0438\u043D\u0432\u0435\u0441\u0442\u043E\u0440\u043E\u0432", n, e, N.PROCEEDS_FROM_INVESTORS), Ue("\u041F\u0440\u043E\u0447\u0438\u0435 \u043F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F", n, e, N.OTHER_PROCEEDS), No("\u0418\u0442\u043E\u0433\u043E \u043F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F", 5, 10), Wa("\u0412\u044B\u043F\u043B\u0430\u0442\u044B"), Ue("\u0412\u044B\u043F\u043B\u0430\u0442\u044B \u0441\u043E\u0442\u0440\u0443\u0434\u043D\u0438\u043A\u0430\u043C (\u0441 \u041D\u0414\u0424\u041B)", n, e, N.EMPLOYEE_PAYMENTS), Ue("\u0412\u044B\u043F\u043B\u0430\u0442\u044B \u043D\u0430\u043B\u043E\u0433\u043E\u0432 \u0441 \u0424\u041E\u0422 (\u0431\u0435\u0437 \u041D\u0414\u0424\u041B)", n, e, N.TAXES_FROM_PAYROLL), Ue("\u041F\u043E\u043A\u0443\u043F\u043A\u0430 \u0442\u043E\u0432\u0430\u0440\u043E\u0432 \u0434\u043B\u044F \u043F\u0435\u0440\u0435\u043F\u0440\u043E\u0434\u0430\u0436\u0438", n, e, N.PURCHASE_OF_RESALE_EQUIPMENT), Ue("\u0410\u0440\u0435\u043D\u0434\u0430 IT \u0438\u043D\u0444\u0440\u0430\u0441\u0442\u0440\u0443\u043A\u0442\u0443\u0440\u044B", n, e, N.IT_INFRASTRUCTURE_RENT), Ue("\u0420\u0435\u0433\u0443\u043B\u044F\u0440\u043D\u044B\u0435 \u043F\u043B\u0430\u0442\u0435\u0436\u0438 \u043F\u043E\u0441\u0442\u0430\u0432\u0449\u0438\u043A\u0430\u043C", n, e, N.REGULAR_PAYMENTS_TO_SUPPLIERS), Ue("\u041F\u0440\u043E\u0447\u0438\u0435 \u043F\u043B\u0430\u0442\u0435\u0436\u0438 \u043F\u043E\u0441\u0442\u0430\u0432\u0449\u0438\u043A\u0430\u043C", n, e, N.OTHER_PAYMENTS_TO_SUPPLIERS), Ue("\u0412\u044B\u043F\u043B\u0430\u0442\u044B \u043F\u0440\u043E\u0446\u0435\u043D\u0442\u043E\u0432 \u043F\u043E \u043A\u0440\u0435\u0434\u0438\u0442\u0430\u043C", n, e, N.INTEREST_PAYMENTS_ON_LOANS), Ue("\u041F\u043E\u0433\u0430\u0448\u0435\u043D\u0438\u0435 \u0442\u0435\u043B\u0430 \u0434\u043E\u043B\u0433\u0430", n, e, N.REPAYMENT_OF_DEBT_BODY), Ue("\u0412\u044B\u043F\u043B\u0430\u0442\u044B \u043D\u0430\u043B\u043E\u0433\u043E\u0432 (\u043A\u0440\u043E\u043C\u0435 \u043D\u0430\u043B\u043E\u0433\u043E\u0432 \u0441 \u0424\u041E\u0422)", n, e, N.TAXES), Ue("\u0412\u044B\u043F\u043B\u0430\u0442\u044B \u0434\u0438\u0432\u0438\u0434\u0435\u043D\u0434\u043E\u0432", n, e, N.DIVIDEND_PAYOUTS), Ue("\u041F\u0440\u043E\u0447\u0438\u0435 \u0432\u044B\u043F\u043B\u0430\u0442\u044B", n, e, N.OTHER_PAYMENTS), No("\u0418\u0442\u043E\u0433\u043E \u0432\u044B\u043F\u043B\u0430\u0442\u044B", 13, 23), Wa("\u041D\u0430\u043A\u043E\u043F\u043B\u0435\u043D\u043D\u044B\u0439 \u0434\u043E\u043B\u0433"), xd(n), hd()];
}
  , Pd = ({year: e, onChange: t, applicationData: a}) => {
    const [n,o] = g.useState([])
      , [i,l] = g.useState([])
      , [s,c] = g.useState([])
      , {isEditable: d} = g.useContext(Lt);
    Bn( () => {
        l(Ad(e, d)),
        c(kd()),
        o(Td(e, (a == null ? void 0 : a.content.economics) || []));
    }
    );
    const p = (u, f) => {
        t == null ? void 0 : t(u, f, e);
    }
    ;
    return r.jsx(K, {
        "data-testid": `economics_${e}`,
        width: "100%",
        overflow: "auto",
        sx: {
            px: 0
        },
        className: "ag-theme-material",
        children: r.jsx(xr, {
            "data-testid": `economics_${e}`,
            id: `economics_${e}`,
            isEditable: d,
            data: n,
            columns: i,
            cell: s,
            activeHeaderClassName: "ag-header-active",
            afterChange: p,
            sheetName: `sheet_${e}`
        })
    });
}
  , Od = () => r.jsxs(K, {
    children: [r.jsx(F, {
        children: "\u042D\u043A\u043E\u043D\u043E\u043C\u0438\u043A\u0430 \u043F\u0440\u043E\u0435\u043A\u0442\u0430 \u0437\u0430\u043F\u043E\u043B\u043D\u044F\u0435\u0442\u0441\u044F \u043A\u0430\u0441\u0441\u043E\u0432\u044B\u043C \u043C\u0435\u0442\u043E\u0434\u043E\u043C. \u041D\u0430 \u044D\u0442\u043E\u043C \u044D\u0442\u0430\u043F\u0435 \u043D\u0430\u0441 \u0438\u043D\u0442\u0435\u0440\u0435\u0441\u0443\u044E\u0442 \u0438\u0441\u043A\u043B\u044E\u0447\u0438\u0442\u0435\u043B\u044C\u043D\u043E \u043F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F \u0438 \u0432\u044B\u043F\u043B\u0430\u0442\u044B \u043F\u043E \u0440\u0430\u0441\u0447\u0435\u0442\u043D\u043E\u043C\u0443 \u0441\u0447\u0435\u0442\u0443 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438."
    }), r.jsxs(F, {
        mt: 1,
        children: [r.jsx(F, {
            component: "span",
            color: "success",
            variant: "strong",
            mr: 1,
            children: "\u0412\u044B\u043F\u043B\u0430\u0442\u044B"
        }), "\u0437\u0430\u043F\u043E\u043B\u043D\u044F\u044E\u0442\u0441\u044F \u0441 \u043F\u043E\u043B\u043E\u0436\u0438\u0442\u0435\u043B\u044C\u043D\u044B\u043C \u0437\u043D\u0430\u043A\u043E\u043C. \u041C\u044B \u0441\u0430\u043C\u0438 \u0432\u044B\u0447\u0442\u0435\u043C \u0438\u0445 \u0438\u0437 \u0432\u0430\u0448\u0438\u0445 \u0434\u043E\u0445\u043E\u0434\u043E\u0432."]
    }), r.jsxs(F, {
        mt: 1,
        children: ["\u041F\u0440\u0438 \u0437\u0430\u043F\u043E\u043B\u043D\u0435\u043D\u0438\u0438 \u0442\u0430\u0431\u043B\u0438\u0446\u044B, \u0441\u0432\u0435\u0440\u044F\u0439\u0442\u0435\u0441\u044C \u0441\u043E \u0441\u0442\u0440\u043E\u043A\u0430\u043C\u0438", r.jsx(F, {
            component: "span",
            color: "success",
            variant: "strong",
            mx: 1,
            children: "\u0414\u0435\u043D\u0435\u0436\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043A\u043E\u043D\u0435\u0446 \u043F\u0435\u0440\u0438\u043E\u0434\u0430"
        }), "\u0438", r.jsx(F, {
            component: "span",
            color: "success",
            variant: "strong",
            ml: 1,
            children: "\u0417\u0430\u0435\u043C\u043D\u044B\u0435 \u0441\u0440\u0435\u0434\u0441\u0442\u0432\u0430 \u043D\u0430 \u043A\u043E\u043D\u0435\u0446 \u043F\u0435\u0440\u0438\u043E\u0434\u0430"
        }), ". \u041E\u043D\u0438 \u0434\u043E\u043B\u0436\u043D\u044B \u0441\u043E\u0432\u043F\u0430\u0434\u0430\u0442\u044C \u0441 \u0434\u0430\u043D\u043D\u044B\u043C\u0438 \u0432\u0430\u0448\u0435\u0439 \u0431\u0443\u0445\u0433\u0430\u043B\u0442\u0435\u0440\u0438\u0438."]
    })]
})
  , jd = ({value: e, onChange: t, years: a, onFormulasValuesUpdate: n, applicationData: o}) => r.jsxs(K, {
    component: "section",
    children: [r.jsx(oa, {
        title: "\u042D\u043A\u043E\u043D\u043E\u043C\u0438\u043A\u0430 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportTitle: "\u042D\u043A\u043E\u043D\u043E\u043C\u0438\u043A\u0430 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportContent: r.jsx(Od, {})
    }), r.jsxs(sd, {
        value: e,
        children: [r.jsx(cd, {
            onChange: t,
            "aria-label": "lab API tabs example",
            children: a.map( (i, l) => r.jsx(Ms, {
                label: i,
                value: String(l + 1)
            }, i))
        }), a.map( (i, l) => r.jsx(md, {
            value: String(l + 1),
            sx: {
                px: 0
            },
            keepMounted: true,
            children: r.jsx(Pd, {
                year: i,
                onChange: n,
                applicationData: o
            })
        }, i))]
    })]
})
  , Id = () => {
    const [e,t] = g.useState("1")
      , [a,n] = g.useState([])
      , {values: o} = Ve()
      , {applicationId: i} = Fe({
        from: Ie
    })
      , {data: l} = Nt(i)
      , {mutate: s} = zt(i)
      , c = g.useCallback( (p, u) => {
        p.preventDefault(),
        t(u);
    }
    , [])
      , d = g.useCallback(async (p, u, f) => {
        if (l) {
            const m = l == null ? void 0 : l.content
              , b = m ? [...m.economics] : [];
            let y = false;
            if (p && ["edit", "Autofill.fill", "CopyPaste.paste"].includes(u)) {
                for (const w of p) {
                    const E = w[0]
                      , A = w[1]
                      , C = w[2]
                      , S = w[3]
                      , T = z.find(gt, {
                        row: E
                    });
                    if (T) {
                        if (z.isNil(C) && z.isNumber(S)) {
                            const P = Ot.utc(f, z.toNumber(A), 1).endOf("month").toISODate();
                            y = true,
                            b.push({
                                year: f,
                                month: z.toNumber(A),
                                key: T.key,
                                value: S,
                                synthetic_date: P
                            });
                        }
                        if (z.isNumber(C) && z.isNumber(S) && C !== S) {
                            y = true;
                            const P = z.findIndex(b, {
                                year: f,
                                month: z.toNumber(A),
                                key: T.key
                            });
                            b[P] = {
                                year: f,
                                month: z.toNumber(A),
                                key: T.key,
                                value: S
                            };
                        }
                        z.isNil(S) && z.isNumber(C) && (y = true,
                        z.remove(b, {
                            year: f,
                            month: z.toNumber(A),
                            key: T.key
                        }));
                    }
                }
                y && s({
                    ...o,
                    economics: b,
                    reason: "economics_changed"
                });
            }
        }
    }
    , [l, s, o]);
    return g.useEffect( () => {
        if (l) {
            const p = Ot.now().year
              , u = Ot.fromISO(l.created_at)
              , f = [];
            f.push(u.year - 1);
            for (let {year: m} = u; m <= p; m += 1)
                f.push(m);
            n(f);
        }
    }
    , [l]),
    r.jsx(jd, {
        applicationData: l,
        value: e,
        onFormulasValuesUpdate: d,
        years: a,
        onChange: c
    });
}
  , Bi = "application_description"
  , _d = ({isEditable: e, onEditSave: t}) => r.jsxs(K, {
    component: "section",
    children: [r.jsx(oa, {
        title: "\u041A\u0440\u0430\u0442\u043A\u043E\u0435 \u043E\u043F\u0438\u0441\u0430\u043D\u0438\u0435 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportTitle: "\u041E\u043F\u0438\u0441\u0430\u043D\u0438\u0435 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportContent: "\u0412\u044B \u0437\u043D\u0430\u0435\u0442\u0435, \u0447\u0442\u043E \u0442\u0430\u043A\u043E\u0435 elevator pitch? \u0421\u0443\u0442\u0435\u0432\u043E\u0439 \u0440\u0430\u0441\u0441\u043A\u0430\u0437 \u043E \u0432\u0430\u0448\u0435\u043C \u043F\u0440\u043E\u0435\u043A\u0442\u0435, \u043A\u043E\u0442\u043E\u0440\u044B\u0439 \u043C\u043E\u0436\u043D\u043E \u0440\u0430\u0441\u0441\u043A\u0430\u0437\u0430\u0442\u044C \u0437\u0430 \u0432\u0440\u0435\u043C\u044F, \u043A\u043E\u0442\u043E\u0440\u043E\u0435 \u043D\u0443\u0436\u043D\u043E \u043B\u0438\u0444\u0442\u0443, \u0447\u0442\u043E\u0431\u044B \u043F\u043E\u0434\u043D\u044F\u0442\u044C\u0441\u044F \u043D\u0430 \u0441\u0430\u043C\u044B\u0439 \u0432\u044B\u0441\u043E\u043A\u0438\u0439 \u044D\u0442\u0430\u0436 \u043D\u0435\u0431\u043E\u0441\u043A\u0440\u0435\u0431\u0430. \u041C\u043E\u0436\u0435\u0442\u0435 \u0432 \u041C\u043E\u0441\u043A\u0432\u0430-\u0421\u0438\u0442\u0438 \u043F\u043E\u0442\u0440\u0435\u043D\u0438\u0440\u043E\u0432\u0430\u0442\u044C\u0441\u044F. \u041D\u043E \u043F\u043E\u0441\u043B\u0435 \u043F\u0440\u043E\u0441\u043B\u0443\u0448\u0438\u0432\u0430\u043D\u0438\u044F \u044D\u0442\u043E\u0433\u043E \u0440\u0430\u0441\u0441\u043A\u0430\u0437\u0430 \u0434\u043E\u043B\u0436\u043D\u043E \u0431\u044B\u0442\u044C \u043F\u043E\u043D\u044F\u0442\u043D\u043E, \u0432 \u0447\u0435\u043C \u0432\u0430\u0448\u0430 \u0438\u0434\u0435\u044F, \u0434\u043B\u044F \u043A\u043E\u0433\u043E \u044D\u0442\u043E \u043D\u0443\u0436\u043D\u043E \u0438 \u0432 \u0447\u0435\u043C \u043E\u0441\u043E\u0431\u0435\u043D\u043D\u043E\u0441\u0442\u044C \u0432\u0430\u0448\u0435\u0433\u043E \u0440\u0435\u0448\u0435\u043D\u0438\u044F. \u0410 \u0435\u0449\u0435 \u043C\u044B \u0431\u044B \u0445\u043E\u0442\u0435\u043B\u0438 \u0432\u0438\u0434\u0435\u0442\u044C \u043F\u0443\u0431\u043B\u0438\u0447\u043D\u044B\u0435 \u0441\u0441\u044B\u043B\u043A\u0438, \u0433\u0434\u0435 \u043D\u0430\u0445\u043E\u0434\u0438\u0442\u0441\u044F \u0432\u0430\u0448\u0435 \u043F\u0440\u0435\u0434\u043B\u043E\u0436\u0435\u043D\u0438\u0435 \u043A\u043B\u0438\u0435\u043D\u0442\u0430\u043C. \u042D\u0442\u043E \u043C\u043E\u0436\u0435\u0442 \u0431\u044B\u0442\u044C \u043F\u0440\u043E\u0441\u0442\u043E \u0441\u0430\u0439\u0442 \u0432\u0430\u0448\u0435\u0433\u043E \u043F\u0440\u043E\u0435\u043A\u0442\u0430, \u0438\u043B\u0438 \u043F\u043E\u0441\u0430\u0434\u043E\u0447\u043D\u044B\u0435 \u0441\u0442\u0440\u0430\u043D\u0438\u0446\u044B \u043A\u043E\u043D\u043A\u0440\u0435\u0442\u043D\u044B\u0445 \u043F\u0440\u043E\u0434\u0443\u043A\u0442\u043E\u0432 - \u043D\u0430\u043C \u0431\u0435\u0437 \u0440\u0430\u0437\u043D\u0438\u0446\u044B."
    }), r.jsx(Vp, {
        maxSize: 1e3,
        fieldName: Bi,
        placeholder: "\u041E\u043F\u0438\u0448\u0438\u0442\u0435, \u0432 \u0447\u0435\u043C \u0441\u0443\u0442\u044C \u0432\u0430\u0448\u0435\u0433\u043E \u043F\u0440\u043E\u0435\u043A\u0442\u0430.",
        onEditSave: t,
        isEditable: e
    })]
})
  , Rd = () => {
    const [e,t] = Ft(Bi)
      , {applicationId: a} = Fe({
        from: Ie
    })
      , {values: n} = Ve()
      , {mutate: o} = zt(a)
      , {isEditable: i} = g.useContext(Lt)
      , l = g.useCallback(async () => {
        e.value.trim() !== t.initialValue && o({
            ...n,
            reason: "description_changed"
        });
    }
    , [e.value, t.initialValue, o, n]);
    return r.jsx(_d, {
        isEditable: i,
        onEditSave: l
    });
}
  , jn = "application_form"
  , Ni = be().required("\u0423\u043A\u0430\u0436\u0438\u0442\u0435 \u0430\u0434\u0440\u0435\u0441 \u0441\u0430\u0439\u0442\u0430").test({
    name: "Url without slash",
    test: (e) => !e.endsWith("/"),
    message: "\u0423\u043A\u0430\u0436\u0438\u0442\u0435 \u0430\u0434\u0440\u0435\u0441 \u0441\u0430\u0439\u0442\u0430 \u0431\u0435\u0437 \u0441\u043B\u0435\u0448\u0430 \u0432 \u043A\u043E\u043D\u0446\u0435"
}).matches(rd, "\u0410\u0434\u0440\u0435\u0441 \u0434\u043E\u043B\u0436\u0435\u043D \u0431\u044B\u0442\u044C \u0432 \u0444\u043E\u0440\u043C\u0430\u0442\u0435 www.example.com")
  , Kt = {
    WEB: "WEB",
    LP: "LP",
    OT: "OT"
}
  , zi = De({
    display_name: be().required("\u0427\u0442\u043E \u043D\u0430\u0445\u043E\u0434\u0438\u0442\u0441\u044F \u043F\u043E \u0441\u0441\u044B\u043B\u043A\u0435?"),
    value: ii().oneOf([...Object.entries(Kt)].map( ([,e]) => e)).defined()
})
  , Dd = (e) => De({
    application_name: be().trim().required("\u041D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u043F\u0440\u043E\u0435\u043A\u0442\u0430 \u043D\u0435 \u043C\u043E\u0436\u0435\u0442 \u0431\u044B\u0442\u044C \u043F\u0443\u0441\u0442\u044B\u043C"),
    application_description: be().max(1e3),
    urls: Ga().of(De().shape({
        url: Ni,
        type: zi
    })),
    files: Ga().of(De().shape({
        file_id: be(),
        type: De().shape({
            display_name: be(),
            value: be()
        })
    })).test("has-investment-presentation", "\u0414\u043E\u0431\u0430\u0432\u044C\u0442\u0435 \u043A\u0430\u043A \u043C\u0438\u043D\u0438\u043C\u0443\u043C \u043E\u0434\u043D\u0443 \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u0443\u044E \u043F\u0440\u0435\u0437\u0435\u043D\u0442\u0430\u0446\u0438\u044E", async (t) => z.some(t, {
        type: {
            value: "IP"
        }
    })),
    investment_requested: za().required("\u041D\u0435\u043E\u0431\u0445\u043E\u0434\u0438\u043C\u043E \u0443\u043A\u0430\u0437\u0430\u0442\u044C \u043F\u0440\u0438\u0432\u043B\u0435\u043A\u0430\u0435\u043C\u0443\u044E \u0441\u0443\u043C\u043C\u0443, \u0435\u0441\u043B\u0438 \u043D\u0435 \u0443\u0432\u0435\u0440\u0435\u043D\u044B, \u0442\u043E \u0432\u0432\u0435\u0434\u0438\u0442\u0435 0"),
    legal_entity: De().test("legal-entity-required", "\u0418\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044F \u043E \u044E\u0440\u0438\u0434\u0438\u0447\u0435\u0441\u043A\u043E\u043C \u043B\u0438\u0446\u0435 \u043D\u0435 \u043F\u0440\u0435\u0434\u043E\u0441\u0442\u0430\u0432\u043B\u0435\u043D\u0430", (t, a) => {
        const {no_legal_entity_reason: n} = a.parent;
        return (n == null ? void 0 : n.value) === sa.NOT_SET ? t && Object.keys(t).length > 0 : true;
    }
    ).when([], {
        is: () => (e == null ? void 0 : e.no_legal_entity_reason) === sa.NOT_SET,
        then: () => De().shape({
            id: be().required(),
            inn: be().required("\u0418\u041D\u041D \u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u0435\u043D"),
            ogrn: be().required("\u041E\u0413\u0420\u041D \u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u0435\u043D"),
            short_name_native: be(),
            long_name_native: be()
        }),
        otherwise: () => De().nullable()
    }),
    no_legal_entity_reason: De().test("reason-required-if-no-legal-entity", "\u041D\u0435\u043E\u0431\u0445\u043E\u0434\u0438\u043C\u043E \u0432\u044B\u0431\u0440\u0430\u0442\u044C \u043F\u0440\u0438\u0447\u0438\u043D\u0443 \u043D\u0435\u043F\u0440\u0435\u0434\u043E\u0441\u0442\u0430\u0432\u043B\u0435\u043D\u0438\u044F \u0434\u0430\u043D\u043D\u044B\u0445 \u043E \u044E\u0440\u0438\u0434\u0438\u0447\u0435\u0441\u043A\u043E\u043C \u043B\u0438\u0446\u0435", (t, a) => {
        const {legal_entity: n} = a.parent;
        return !n || Object.keys(n).length === 0 ? t && t.value !== sa.NOT_SET : true;
    }
    ).when([], {
        is: () => z.isNil(e == null ? void 0 : e.legal_entity),
        then: () => De().shape({
            display_name: be(),
            value: be()
        }),
        otherwise: () => De().nullable()
    }),
    economics: Ga().required().of(De().shape({
        key: be().required(),
        year: za().required(),
        month: za().required(),
        value: za().optional()
    })).test("revenue-field", '\u0414\u043B\u044F \u043F\u043E\u0434\u0430\u0447\u0438 \u0437\u0430\u044F\u0432\u043A\u0438 \u0434\u043E\u043B\u0436\u043D\u0430 \u0431\u044B\u0442\u044C \u0437\u0430\u043F\u043E\u043B\u043D\u0435\u043D\u0430 \u0445\u043E\u0442\u044F \u0431\u044B \u0441\u0442\u0440\u043E\u043A\u0430 "\u041F\u043E\u0441\u0442\u0443\u043F\u043B\u0435\u043D\u0438\u044F \u043E\u0442 \u043A\u043B\u0438\u0435\u043D\u0442\u043E\u0432 \u0441 \u043F\u0440\u043E\u0434\u0430\u0436"', (t) => {
        const a = Ot.now()
          , n = e ? Ot.fromISO(e.created_at) : void 0;
        if (n) {
            const o = Ot.fromObject({
                year: n.year - 1,
                month: 1,
                day: 1
            })
              , i = a.diff(o, ["months"])
              , l = Math.trunc(i.months);
            return z.filter(t, {
                key: N.PROCEEDS_FROM_CUSTOMER_SALES
            }).length === l;
        }
        return true;
    }
    )
})
  , Fd = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = Nt(e)
      , {data: a} = Hn()
      , {data: n} = ti(e)
      , o = g.useMemo( () => {
        const s = t == null ? void 0 : t.no_legal_entity_reason;
        return s === sa.NOT_SET ? {
            display_name: "",
            value: ""
        } : ka(s) ? z.find(a == null ? void 0 : a.no_legal_entity_reason, (c) => c.value === s) : {
            display_name: "",
            value: ""
        };
    }
    , [t == null ? void 0 : t.no_legal_entity_reason, a == null ? void 0 : a.no_legal_entity_reason])
      , i = g.useMemo( () => n == null ? void 0 : n.map( (s) => ({
        file_id: s.id,
        type: {
            display_name: "should not be displayed",
            value: s.file_type
        }
    })), [n])
      , l = (t == null ? void 0 : t.investment_requested) && z.toNumber(t == null ? void 0 : t.investment_requested);
    return g.useMemo( () => ({
        application_name: (t == null ? void 0 : t.name) || "",
        application_description: (t == null ? void 0 : t.description) || "",
        investment_requested: l,
        legal_entity: t == null ? void 0 : t.legal_entity,
        no_legal_entity_reason: o,
        economics: (t == null ? void 0 : t.content.economics) || [],
        files: i || [],
        reason: ""
    }), [t == null ? void 0 : t.name, t == null ? void 0 : t.description, t == null ? void 0 : t.legal_entity, t == null ? void 0 : t.content.economics, l, o, i]);
}
  , Ld = ({id: e="", applicationId: t="", isDeletable: a}) => {
    const {mutate: n, isPending: o, isSuccess: i} = Er(t, e)
      , l = g.useCallback( () => {
        n();
    }
    , [n]);
    return r.jsx(yp, {
        id: e,
        onDelete: l,
        isDeletable: a,
        isDisabled: o || i
    });
}
  , ft = {
    IP: "IP",
    PP: "PP",
    BP: "BP",
    IT: "IT",
    OTP: "OTP",
    CRM: "CRM",
    BS: "BS",
    PL: "PL",
    CF: "CF",
    FM: "FM",
    OT: "OT"
}
  , Md = () => {
    const {values: e} = Ve()
      , t = g.useMemo( () => z.some(e.files, {
        type: {
            value: ft.IP
        }
    }), [e.files]);
    return r.jsxs(K, {
        display: "flex",
        children: [r.jsxs(F, {
            ml: 2,
            children: ["\u0414\u043E\u0431\u0430\u0432\u044C\u0442\u0435 \u043A\u0430\u043A \u043C\u0438\u043D\u0438\u043C\u0443\u043C", " ", r.jsx(F, {
                variant: "strong",
                ml: 0.1,
                color: "warning",
                component: "span",
                children: "\u0418\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u0443\u044E \u043F\u0440\u0435\u0437\u0435\u043D\u0442\u0430\u0446\u0438\u044E"
            })]
        }), t ? r.jsxs(F, {
            ml: 2,
            display: "flex",
            color: "textDisabled",
            children: [r.jsx(F, {
                component: "span",
                children: "(\u043E\u0442\u043B\u0438\u0447\u043D\u043E, \u0444\u0430\u0439\u043B \u0435\u0441\u0442\u044C"
            }), r.jsx(Wt, {
                icon: je.check,
                color: "success"
            }), ")"]
        }) : r.jsxs(F, {
            ml: 2,
            display: "flex",
            color: "textDisabled",
            children: [r.jsx(F, {
                component: "span",
                children: "(\u043F\u043E\u043A\u0430 \u043D\u0438 \u043E\u0434\u043D\u043E\u0439 \u043D\u0435 \u043D\u0430\u0439\u0434\u0435\u043D\u043E"
            }), r.jsx(Wt, {
                icon: je.cross,
                color: "error"
            }), ")"]
        })]
    });
}
  , $d = [ft.IP, ft.PP, ft.BP, ft.IT, ft.OTP, ft.CRM, ft.BS, ft.PL, ft.CF, ft.FM, ft.OT]
  , Bd = ({fileFiled: e, index: t, isLoading: a, errorMessage: n, onDelete: o}) => {
    const i = va()
      , {data: l} = Ii()
      , s = `${e.file.name.split(".")[0]}_${Ta()}`
      , c = g.useMemo( () => {
        const p = l.file_type;
        return z.filter(p, (u) => $d.includes(u.value));
    }
    , [l])
      , d = g.useMemo( () => e.file.name.length > 28 ? `${e.file.name.slice(0, 27)}...` : e.file.name, [e.file.name]);
    return r.jsxs(Ia, {
        "data-testid": `${s}_field`,
        variant: "outlined",
        sx: {
            display: "flex",
            flexDirection: "column"
        },
        children: [r.jsx(_a, {
            sx: {
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                height: 200
            },
            children: r.jsx(Bp, {
                fieldName: `files.${t}.file`,
                fileType: e.file.type
            })
        }), r.jsxs(Zn, {
            sx: {
                display: "flex",
                flexDirection: "column",
                mt: "auto"
            },
            children: [r.jsxs(K, {
                sx: {
                    width: "100%",
                    justifyContent: "start",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    wordWrap: "break-word"
                },
                children: [r.jsx(od, {
                    fieldName: `files.${t}.type`,
                    choices: c,
                    emptyLabel: "\u0422\u0438\u043F \u0434\u043E\u043A\u0443\u043C\u0435\u043D\u0442\u0430 \u043D\u0435 \u0443\u043A\u0430\u0437\u0430\u043D",
                    emptyColor: n ? "error" : "primary",
                    emptyVariant: "outlined"
                }), n && r.jsx(Qo, {
                    error: true,
                    children: r.jsx(Zo, {
                        children: `${n}`
                    })
                }, Ta())]
            }), r.jsxs(K, {
                display: "flex",
                width: "100%",
                justifyContent: "space-between",
                alignItems: "center",
                children: [r.jsxs(K, {
                    children: [r.jsx(F, {
                        fontSize: "0.8rem",
                        children: d
                    }), r.jsx(F, {
                        variant: "small",
                        color: "textSecondary",
                        children: `${i(e.file.size).format("0.0 b")}`
                    })]
                }), r.jsx(Jn, {
                    isDisabled: a,
                    sx: {
                        fontSize: "1.5rem"
                    },
                    variant: "thin",
                    onClick: o(t)
                })]
            })]
        })]
    });
}
  , Nd = ({isLoading: e}) => {
    const {values: t, errors: a} = Ve()
      , n = g.useRef(null)
      , [o,i] = g.useState([])
      , l = g.useCallback( (c) => {
        if (n.current) {
            const d = [];
            for (const p of c) {
                n.current.push({
                    file: p,
                    type: {}
                });
                const u = Ta();
                d.push(u);
            }
            i([...o, ...d]);
        }
    }
    , [o])
      , s = (c) => (d) => {
        if (d.preventDefault(),
        n.current) {
            n.current.remove(c);
            const p = [...o];
            p.splice(c, 1),
            i(p);
        }
    }
    ;
    return r.jsxs(r.Fragment, {
        children: [r.jsx(fp, {
            filesLimit: 10,
            onDrop: l,
            dropzoneParagraphStyles: {
                fontSize: "0.95rem"
            },
            dropZoneStyles: {
                minHeight: "150px"
            },
            dropzoneText: "\u041F\u0435\u0440\u0435\u0442\u0430\u0449\u0438\u0435 \u0441\u044E\u0434\u0430 \u0444\u0430\u0439\u043B\u044B \u0438\u043B\u0438 \u043A\u043B\u0438\u043A\u043D\u0438\u0442\u0435 \u0434\u043B\u044F \u0434\u043E\u0431\u0430\u0432\u043B\u0435\u043D\u0438\u044F",
            accept: {
                "application/pdf": [".pdf"],
                "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
                "application/msword": [".doc"],
                "applicatiapplication/vnd.openxmlformats-officedocument.presentationml.presentation": [".pptx"],
                "application/vnd.ms-powerpoint": [".ppt"],
                "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [".xlsx"],
                "application/vnd.ms-excel": [".xls"]
            }
        }), r.jsx(K, {
            sx: {
                display: "grid",
                rowGap: 1,
                columnGap: 1,
                gridTemplateColumns: "repeat(2, minmax(250px, 1fr));",
                justifyContent: "start",
                mt: 2
            },
            children: r.jsx(Or, {
                name: "files",
                children: (c) => {
                    var _a2;
                    return n.current = c,
                    (_a2 = t.files) == null ? void 0 : _a2.map( (d, p) => {
                        var _a3;
                        let u = "";
                        return ka(a == null ? void 0 : a.files) && (u = ((_a3 = a.files[p]) == null ? void 0 : _a3.type.display_name) || ""),
                        r.jsx(Bd, {
                            fileFiled: d,
                            index: p,
                            isLoading: e,
                            onDelete: s,
                            errorMessage: u
                        }, Ta());
                    }
                    );
                }
            })
        })]
    });
}
  , In = "new_file_form"
  , zd = De().shape({
    files: Ga().min(1).of(De({
        file: ii(),
        type: De().shape({
            display_name: be().required("\u0412\u044B\u0431\u0435\u0440\u0438\u0442\u0435 \u0442\u0438\u043F \u0434\u043E\u043A\u0443\u043C\u0435\u043D\u0442\u0430"),
            value: be().defined()
        })
    }))
})
  , Vd = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {submit: t, isPending: a} = ai();
    return r.jsx(Ra, {
        initialValues: {
            files: []
        },
        enableReinitialize: true,
        onSubmit: (n) => t({
            ...n,
            application_id: e
        }),
        validationSchema: zd,
        children: ({handleSubmit: n}) => r.jsx(Da, {
            onSubmit: n,
            name: In,
            id: In,
            children: r.jsx(Nd, {
                isLoading: a
            })
        })
    });
}
  , Hd = ({showModal: e, setShowModal: t}) => {
    const {isPending: a, isSuccess: n, data: o} = ai()
      , i = g.useCallback(async () => {
        t(false);
    }
    , [t]);
    return g.useEffect( () => {
        n && o && t(false);
    }
    , [o, n, t]),
    r.jsx(on, {
        title: "\u0417\u0430\u0433\u0440\u0443\u0437\u0438\u0442\u044C \u0434\u043E\u043A\u0443\u043C\u0435\u043D\u0442\u044B \u043F\u043E \u043F\u0440\u043E\u0435\u043A\u0442\u0443",
        content: r.jsx(Vd, {}),
        showModal: e,
        acceptButtonLabel: "\u0417\u0430\u0433\u0440\u0443\u0437\u0438\u0442\u044C",
        refuseButtonLabel: "\u041E\u0442\u043C\u0435\u043D\u0438\u0442\u044C",
        onRefuse: i,
        acceptButtonColor: "success",
        refuseButtonColor: "error",
        isLoading: a,
        acceptButtonFormName: In,
        acceptButtonType: "submit"
    });
}
  , Ud = ({applicationFiles: e=[], applicationId: t="", isEditable: a=true, touched: n, errors: o, showModal: i, setShowModal: l, onFileAdd: s}) => r.jsxs(K, {
    component: "section",
    sx: {
        mt: 3
    },
    children: [r.jsx(ca, {
        container: true,
        spacing: 2,
        children: e.map( (c) => r.jsx(ca, {
            size: 12,
            children: r.jsx(Ld, {
                applicationId: t,
                id: c.id,
                isDeletable: a
            })
        }, c.id))
    }), a && r.jsxs(r.Fragment, {
        children: [r.jsx($n, {
            onClick: s,
            color: n.files && o.files ? "error" : "primary",
            variant: n.files && o.files ? "outlined" : "text",
            children: "+ \u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u0444\u0430\u0439\u043B\u044B"
        }), r.jsx(Md, {
            applicationId: t
        })]
    }), r.jsx(Hd, {
        showModal: i,
        setShowModal: l
    })]
})
  , qd = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = ti(e)
      , {isEditable: a} = g.useContext(Lt)
      , {errors: n, touched: o, setTouched: i} = Ve()
      , [l,s] = g.useState(false)
      , c = g.useCallback( () => {
        s(true),
        i({
            ...o,
            files: true
        });
    }
    , [i, o]);
    return r.jsx(Ud, {
        applicationId: e,
        applicationFiles: t,
        touched: o,
        errors: n,
        isEditable: a,
        showModal: l,
        setShowModal: s,
        onFileAdd: c
    });
}
  , qo = "investment_requested"
  , Wd = ({isEditable: e}) => {
    const [t,a] = Ft(qo)
      , {values: n} = Ve()
      , {applicationId: o} = Fe({
        from: Ie
    })
      , {mutate: i} = zt(o)
      , l = g.useCallback(async () => {
        t.value !== a.initialValue && i({
            ...n,
            reason: "changed_investments_requested"
        });
    }
    , [t.value, a.initialValue, i, n]);
    return r.jsx(Yp, {
        label: "\u041A\u0430\u043A\u0443\u044E \u0441\u0443\u043C\u043C\u0443 \u0432\u044B \u043F\u0440\u0438\u0432\u043B\u0435\u043A\u0430\u0435\u0442\u0435?",
        fieldName: qo,
        placeholder: "5 000 000",
        currency: "\u0440\u0443\u0431.",
        isEditable: e,
        onEditSave: l,
        sx: {
            mt: 1,
            width: "100%"
        }
    });
}
  , Yd = ({isEditable: e}) => r.jsxs(K, {
    component: "section",
    children: [r.jsx(oa, {
        title: "\u0418\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u044B\u0435 \u043C\u0430\u0442\u0435\u0440\u0438\u0430\u043B\u044B",
        supportTitle: "\u0418\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u044B\u0435 \u043C\u0430\u0442\u0435\u0440\u0438\u0430\u043B\u044B",
        supportContent: "\u041D\u0430 \u0434\u0430\u043D\u043D\u043E\u043C \u044D\u0442\u0430\u043F\u0435 \u043C\u044B \u043E\u0436\u0438\u0434\u0430\u0435\u043C \u043E\u0442 \u0432\u0430\u0441 \u043A\u0430\u043A\u0438\u0435-\u043B\u0438\u0431\u043E \u0434\u043E\u043A\u0443\u043C\u0435\u043D\u0442\u044B, \u0433\u0434\u0435 \u0438\u0437\u043B\u043E\u0436\u0435\u043D\u043E \u043E\u043F\u0438\u0441\u0430\u043D\u0438\u0435 \u0432\u0430\u0448\u0435\u0433\u043E \u043F\u0440\u043E\u0435\u043A\u0442\u0430 \u043A\u0430\u043A \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u043E\u0433\u043E \u043F\u0440\u0435\u0434\u043B\u043E\u0436\u0435\u043D\u0438\u044F. \u041A\u0430\u043A \u043C\u0438\u043D\u0438\u043C\u0443\u043C, \u043C\u044B \u043E\u0436\u0438\u0434\u0430\u0435\u043C, \u0447\u0442\u043E \u0432\u044B \u0434\u0430\u0434\u0438\u0442\u0435 \u043D\u0430\u043C \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u043E\u043D\u043D\u0443\u044E \u043F\u0440\u0435\u0437\u0435\u043D\u0442\u0430\u0446\u0438\u044E. \u0415\u0441\u043B\u0438 \u0443 \u0432\u0430\u0441 \u0435\u0441\u0442\u044C \u0434\u0440\u0443\u0433\u0438\u0435 \u0434\u043E\u043A\u0443\u043C\u0435\u043D\u0442\u044B, \u043A\u043E\u0442\u043E\u0440\u044B\u0435 \u043C\u043E\u0433\u0443\u0442 \u043F\u043E\u043C\u043E\u0447\u044C \u043D\u0430\u043C \u043B\u0443\u0447\u0448\u0435 \u043F\u043E\u043D\u044F\u0442\u044C \u0432\u0430\u0448 \u043F\u0440\u043E\u0435\u043A\u0442 (\u043F\u0440\u043E\u0434\u0443\u043A\u0442\u043E\u0432\u044B\u0435 \u043F\u0440\u0435\u0437\u0435\u043D\u0442\u0430\u0446\u0438\u0438, \u043A \u043F\u0440\u0438\u043C\u0435\u0440\u0443), \u043F\u0440\u0438\u043A\u043B\u0430\u0434\u044B\u0432\u0430\u0439\u0442\u0435 \u0438\u0445 \u0442\u043E\u0436\u0435."
    }), r.jsx(Wd, {
        isEditable: e
    })]
})
  , Gd = () => {
    const {isEditable: e} = g.useContext(Lt);
    return r.jsx(Yd, {
        isEditable: e
    });
}
  , Kd = (e) => {
    const t = `${zn.legalEntities.searchLegalEntityByInnOrOgrn}?inn-or-ogrn=${e}`;
    return an({
        url: t,
        method: "get"
    });
}
  , Xd = (e) => {
    const t = ei(zn.legalEntities.owners, {
        id: e
    });
    return an({
        url: t,
        method: "get"
    });
}
  , Qd = (e) => {
    const t = ei(zn.legalEntities.beneficiaries, {
        id: e
    });
    return an({
        url: t,
        method: "get"
    });
}
  , Zd = (e) => Vn({
    queryKey: nn.legalEntities.searchByInnOrOgrn(e).queryKey,
    queryFn: () => Kd(e),
    enabled: !!e
})
  , Jd = (e) => Vn({
    queryKey: nn.legalEntities.owners(e).queryKey,
    queryFn: () => Xd(e),
    enabled: !!e
})
  , eu = (e) => Vn({
    queryKey: nn.legalEntities.beneficiaries(e).queryKey,
    queryFn: () => Qd(e),
    enabled: !!e
})
  , tu = (e) => ta(Zd(e))
  , au = (e) => ta(Jd(e))
  , nu = (e) => ta(eu(e));
Cr.registerModules([Sr]);
const Vi = ({isPending: e, rowData: t, colDefs: a, getRowId: n, suppressCellFocus: o, pinnedBottomRowData: i, pagination: l}) => {
    const s = g.useRef(null)
      , c = () => {
        var _a2;
        ((_a2 = s == null ? void 0 : s.current) == null ? void 0 : _a2.api) && s.current.api.sizeColumnsToFit();
    }
    ;
    return g.useEffect( () => {
        var _a2;
        return ((_a2 = s == null ? void 0 : s.current) == null ? void 0 : _a2.api) && setTimeout( () => c(), 100),
        window.addEventListener("resize", c),
        () => {
            window.removeEventListener("resize", c);
        }
        ;
    }
    , []),
    r.jsx(yr, {
        ref: s,
        theme: wr,
        localeText: ip,
        suppressScrollOnNewData: true,
        domLayout: "autoHeight",
        rowData: t,
        columnDefs: a,
        pinnedBottomRowData: i,
        pagination: l,
        paginationPageSizeSelector: [25, 50, 100],
        paginationPageSize: 25,
        loading: e,
        suppressCellFocus: o,
        defaultColDef: {
            flex: 1,
            minWidth: 150,
            autoHeight: true
        },
        autoSizeStrategy: {
            type: "fitGridWidth"
        },
        getRowId: n
    });
}
  , ou = (e, t) => {
    const a = [];
    for (const n of e) {
        const o = {
            name: n.actor_name,
            type: n.actor_type,
            citizenship: n.owner_origin,
            amount: n.ownership_units * t.nominal_price,
            share: n.share
        };
        a.push(o);
    }
    return a;
}
  , iu = (e) => {
    const t = e.reduce( (n, o) => n + o.amount, 0)
      , a = e.reduce( (n, o) => n + o.share, 0);
    return [{
        amount: t,
        share: a
    }];
}
  , ru = ({beneficiaries: e, equity: t}) => {
    const [a,n] = g.useState([])
      , o = xi()
      , i = hi()
      , [l,s] = h.useState([])
      , [c] = h.useState([{
        field: "name",
        headerName: "\u041D\u0430\u0438\u043C\u0435\u043D\u043E\u0432\u0430\u043D\u0438\u0435",
        flex: 1
    }, {
        field: "type",
        headerName: "\u0422\u0438\u043F \u0432\u043B\u0430\u0434\u0435\u043B\u044C\u0446\u0430"
    }, {
        field: "citizenship",
        headerName: "\u0421\u0442\u0440\u0430\u043D\u0430"
    }, {
        field: "amount",
        headerName: "\u0420\u0430\u0437\u043C\u0435\u0440 \u0432 \u0423\u041A, \u0440\u0443\u0431.",
        valueFormatter: o
    }, {
        field: "share",
        headerName: "\u0414\u043E\u043B\u044F, %",
        valueFormatter: i,
        sort: "desc"
    }]);
    return g.useEffect( () => {
        if (e.length > 0) {
            const d = ou(e, t)
              , p = iu(d);
            s(d),
            n(p);
        }
    }
    , [t, t.nominal_price, t.quantity, e]),
    r.jsx(Vi, {
        colDefs: c,
        rowData: l,
        pagination: false,
        suppressCellFocus: true,
        pinnedBottomRowData: a
    });
}
  , lu = ({entity: e}) => {
    const {data: t} = nu(e.id)
      , a = g.useMemo( () => {
        if (lp(t == null ? void 0 : t.owners))
            return false;
        let n = 0;
        for (const o of (t == null ? void 0 : t.owners) || [])
            o.actor_type === "\u0424\u0438\u0437.\u043B\u0438\u0446\u043E" && (n += o.share);
        return n < 1;
    }
    , [t]);
    return t && a && r.jsxs(K, {
        component: "section",
        "data-testid": "owners",
        children: [r.jsx(oa, {
            title: "\u041A\u043E\u043D\u0435\u0447\u043D\u044B\u0435 \u0431\u0435\u043D\u0435\u0444\u0438\u0446\u0438\u0430\u0440\u044B \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438",
            variant: "h3"
        }), r.jsx(ru, {
            beneficiaries: t == null ? void 0 : t.owners,
            equity: t == null ? void 0 : t.asset
        })]
    });
}
  , Hi = ({applicationId: e, buttonLabel: t, variant: a, size: n}) => {
    const {mutate: o} = zt(e)
      , {values: i} = Ve()
      , l = g.useCallback(async () => {
        o({
            ...i,
            reason: "legal_data_reset"
        });
    }
    , [o, i]);
    return r.jsx(bn, {
        onClick: l,
        variant: a,
        size: n,
        children: t
    });
}
  , su = ({companyData: e, applicationId: t, isEditable: a}) => {
    const n = g.useMemo( () => e ? Ot.fromISO(e == null ? void 0 : e.reg_date).setLocale("ru").toLocaleString() : "", [e]);
    return r.jsxs(Ia, {
        variant: "outlined",
        sx: {
            mt: 2
        },
        children: [r.jsxs(_a, {
            children: [r.jsx(F, {
                gutterBottom: true,
                sx: {
                    color: "text.secondary",
                    fontSize: 14
                },
                children: e == null ? void 0 : e.short_name_native
            }), r.jsx(F, {
                sx: {
                    color: "text.secondary",
                    fontSize: 12
                },
                children: `\u0418\u041D\u041D: ${e == null ? void 0 : e.inn}`
            }), r.jsx(F, {
                sx: {
                    color: "text.secondary",
                    fontSize: 12
                },
                children: `\u041E\u0413\u0420\u041D: ${e == null ? void 0 : e.ogrn}`
            }), r.jsx(F, {
                sx: {
                    color: "text.secondary",
                    fontSize: 12
                },
                children: `\u0417\u0430\u0440\u0435\u0433\u0438\u0441\u0442\u0440\u0438\u0440\u043E\u0432\u0430\u043D\u0430: ${n}`
            })]
        }), r.jsx(Zn, {
            children: r.jsx(Hi, {
                buttonLabel: "\u0423\u0434\u0430\u043B\u0438\u0442\u044C \u043F\u0440\u0438\u0432\u044F\u0437\u043A\u0443 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438",
                applicationId: t,
                variant: "text",
                size: "small",
                disabled: !a
            })
        })]
    });
}
  , cu = (e, t) => {
    const a = [];
    for (const n of e) {
        const o = {
            name: n.actor_name,
            type: n.actor_type,
            citizenship: n.owner_origin,
            amount: n.ownership_units * t.nominal_price,
            share: n.share
        };
        a.push(o);
    }
    return a;
}
  , pu = (e) => {
    const t = e.reduce( (n, o) => n + o.amount, 0)
      , a = e.reduce( (n, o) => n + o.share, 0);
    return [{
        amount: t,
        share: a
    }];
}
  , du = ({owners: e, equity: t}) => {
    const [a,n] = g.useState([])
      , o = xi()
      , i = hi()
      , [l,s] = h.useState([])
      , [c] = h.useState([{
        field: "name",
        headerName: "\u041D\u0430\u0438\u043C\u0435\u043D\u043E\u0432\u0430\u043D\u0438\u0435",
        flex: 1
    }, {
        field: "type",
        headerName: "\u0422\u0438\u043F \u0432\u043B\u0430\u0434\u0435\u043B\u044C\u0446\u0430"
    }, {
        field: "citizenship",
        headerName: "\u0421\u0442\u0440\u0430\u043D\u0430"
    }, {
        field: "amount",
        headerName: "\u0420\u0430\u0437\u043C\u0435\u0440 \u0432 \u0423\u041A, \u0440\u0443\u0431.",
        valueFormatter: o
    }, {
        field: "share",
        headerName: "\u0414\u043E\u043B\u044F, %",
        valueFormatter: i,
        sort: "desc"
    }]);
    return g.useEffect( () => {
        if (e.length > 0) {
            const d = cu(e, t)
              , p = pu(d);
            s(d),
            n(p);
        }
    }
    , [t, t.nominal_price, t.quantity, e]),
    r.jsx(Vi, {
        colDefs: c,
        rowData: l,
        pagination: false,
        suppressCellFocus: true,
        pinnedBottomRowData: a
    });
}
  , uu = ({entity: e}) => {
    const {data: t} = au(e.id);
    return t && r.jsxs(K, {
        component: "section",
        "data-testid": "owners",
        children: [r.jsx(oa, {
            title: "\u0412\u043B\u0430\u0434\u0435\u043B\u044C\u0446\u044B \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438",
            variant: "h3"
        }), r.jsx(Qn, {
            label: "\u0418\u0441\u0442\u043E\u0447\u043D\u0438\u043A: \u0424\u041D\u0421 \u0420\u0424",
            color: "success",
            variant: "outlined",
            sx: {
                borderRadius: 2,
                mb: 2
            },
            size: "small"
        }), r.jsx(du, {
            owners: t == null ? void 0 : t.owners,
            equity: t == null ? void 0 : t.asset
        })]
    });
}
  , mu = ({applicationData: e, applicationId: t, isEditable: a}) => r.jsxs(K, {
    "data-testid": "legal_data",
    display: "flex",
    flexDirection: "column",
    width: "100%",
    children: [r.jsx(ca, {
        container: true,
        spacing: 2,
        children: r.jsx(ca, {
            size: 4,
            children: r.jsx(su, {
                companyData: e == null ? void 0 : e.legal_entity,
                applicationId: t,
                isEditable: a
            })
        })
    }), (e == null ? void 0 : e.legal_entity) && r.jsx(uu, {
        entity: e.legal_entity
    }), (e == null ? void 0 : e.legal_entity) && r.jsx(lu, {
        entity: e.legal_entity
    })]
})
  , fu = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = Nt(e)
      , {isEditable: a} = g.useContext(Lt);
    return r.jsx(mu, {
        applicationData: t,
        applicationId: e,
        isEditable: a
    });
}
  , vu = ({isEditable: e, reason: t, applicationId: a}) => r.jsx(K, {
    display: "flex",
    mt: 2,
    children: r.jsxs(Ia, {
        variant: "outlined",
        children: [r.jsxs(_a, {
            children: [r.jsx(F, {
                gutterBottom: true,
                sx: {
                    color: "text.secondary",
                    fontSize: 14
                },
                children: "\u041D\u0435\u0442 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438 \u043F\u0440\u0438\u0432\u044F\u0437\u0430\u043D\u043D\u043E\u0439 \u043A \u0437\u0430\u044F\u0432\u043A\u0435."
            }), r.jsxs(F, {
                sx: {
                    fontSize: 14
                },
                children: ["\u041F\u0440\u0438\u0447\u0438\u043D\u0430:", r.jsx(F, {
                    color: "primary",
                    component: "span",
                    sx: {
                        ml: 0.5,
                        fontSize: 14
                    },
                    children: `${t == null ? void 0 : t.display_name}`
                })]
            })]
        }), r.jsx(Zn, {
            children: r.jsx(Hi, {
                applicationId: a,
                buttonLabel: '\u0423\u0434\u0430\u043B\u0438\u0442\u044C \u0441\u0442\u0430\u0442\u0443\u0441 "\u041D\u0435\u0442 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438"',
                disabled: !e,
                variant: "text",
                size: "small"
            })
        })]
    })
})
  , gu = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {isEditable: t} = g.useContext(Lt)
      , {data: a} = Nt(e)
      , {data: n} = Hn()
      , o = g.useMemo( () => z.find(n == null ? void 0 : n.no_legal_entity_reason, {
        value: a == null ? void 0 : a.no_legal_entity_reason
    }), [a == null ? void 0 : a.no_legal_entity_reason, n == null ? void 0 : n.no_legal_entity_reason]);
    return r.jsx(vu, {
        applicationId: e,
        isEditable: t,
        reason: o
    });
}
  , xu = ({inn: e, regNumber: t, shortName: a, longName: n}) => r.jsxs(K, {
    children: [r.jsx(F, {
        children: a || n
    }), r.jsxs(F, {
        fontSize: 12,
        color: "textDisabled",
        children: ["\u0418\u041D\u041D: ", e]
    }), r.jsxs(F, {
        fontSize: 12,
        color: "textDisabled",
        children: ["\u041E\u0413\u0420\u041D: ", t]
    })]
})
  , hu = () => r.jsxs(K, {
    mb: 2,
    children: [r.jsxs(F, {
        children: ["\u0412 \u043F\u0440\u0435\u0434\u043B\u043E\u0436\u0435\u043D\u043D\u0443\u044E \u0444\u043E\u0440\u043C\u0443 \u0432\u0432\u0435\u0434\u0438\u0442\u0435", r.jsx(F, {
            variant: "strong",
            mx: 0.5,
            color: "success",
            children: "\u0418\u041D\u041D"
        }), "\u0438\u043B\u0438", r.jsx(F, {
            variant: "strong",
            mx: 0.5,
            color: "success",
            children: "\u041E\u0413\u0420\u041D"
        }), "\u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438, \u0432 \u043A\u043E\u0442\u043E\u0440\u0443\u044E \u0432\u044B \u043F\u0440\u0438\u0432\u043B\u0435\u043A\u0430\u0435\u0442\u0435 \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u0438. \u0418\u0437 \u0432\u044B\u043F\u0430\u0434\u0430\u044E\u0449\u0435\u0433\u043E \u0441\u043F\u0438\u0441\u043A\u0430 \u0432\u044B\u0431\u0435\u0440\u0438\u0442\u0435 \u0432\u0430\u0448\u0443 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u044E."]
    }), r.jsxs(F, {
        children: ["\u0415\u0441\u043B\u0438 \u043F\u043E\u0441\u043B\u0435 \u0432\u0432\u043E\u0434\u0430 \u0434\u0430\u043D\u043D\u044B\u0445, \u0432\u044B \u043D\u0435 \u0432\u0438\u0434\u0438\u0442\u0435 \u0432 \u0441\u043F\u0438\u0441\u043A\u0435 \u0441\u0432\u043E\u044E \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u044E, \u0442\u043E \u043F\u0440\u043E\u0432\u0435\u0440\u044C\u0442\u0435 \u0432\u0432\u0435\u0434\u0435\u043D\u043D\u044B\u0435 \u0418\u041D\u041D \u0438\u043B\u0438 \u041E\u0413\u0420\u041D. \u041E\u0431\u0440\u0430\u0442\u0438\u0442\u0435 \u0432\u043D\u0438\u043C\u0430\u043D\u0438\u0435, \u0447\u0442\u043E \u043C\u044B \u0434\u0430\u0435\u043C \u043D\u0430 \u0432\u044B\u0431\u043E\u0440 \u0442\u043E\u043B\u044C\u043A\u043E \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438,", r.jsx(F, {
            variant: "strong",
            color: "primary",
            ml: 0.5,
            children: "\u0441\u043E\u0437\u0434\u0430\u043D\u043D\u044B\u0435 \u0432 \u0444\u043E\u0440\u043C\u0435 \u041E\u041E\u041E \u0438 \u0410\u041E"
        }), ". \u0415\u0441\u043B\u0438 \u0432\u0430\u0448\u0430 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u044F \u0441\u043E\u0437\u0434\u0430\u043D\u0430 \u0432 \u0434\u0440\u0443\u0433\u043E\u0439 \u0444\u043E\u0440\u043C\u0435, \u0442\u043E \u043D\u0430\u0436\u043C\u0438\u0442\u0435 \u043D\u0430 \u043A\u043D\u043E\u043F\u043A\u0443", r.jsx(F, {
            mx: 0.5,
            color: "error",
            component: "span",
            children: "\u041E\u0422\u041C\u0415\u041D\u0418\u0422\u042C"
        }), "\u0438 \u0432\u044B\u0431\u0435\u0440\u0438\u0442\u0435 \u0432\u0430\u0440\u0438\u0430\u043D\u0442", " ", r.jsx(F, {
            ml: 0.5,
            color: "primary",
            component: "span",
            children: "\u041D\u0415\u0422 \u041A\u041E\u041C\u041F\u0410\u041D\u0418\u0418"
        }), "."]
    })]
})
  , _n = "new_company_form"
  , bu = De({
    legal_entity: De().shape({
        id: be().required(),
        inn: be().required("\u0418\u041D\u041D \u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u0435\u043D"),
        ogrn: be().required("\u041E\u0413\u0420\u041D \u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u0435\u043D"),
        short_name_native: be(),
        long_name_native: be()
    })
})
  , yu = ({selectedOption: e, setSelectedOption: t}) => {
    const {initialValues: a} = Ve()
      , [n,o] = g.useState("")
      , {data: i, isLoading: l} = tu(n)
      , {applicationId: s} = Fe({
        from: Ie
    })
      , {mutate: c} = zt(s)
      , d = g.useMemo( () => n.length !== 10 && n.length !== 13 ? "\u041D\u0435\u043E\u0431\u0445\u043E\u0434\u0438\u043C\u043E \u0432\u0432\u0435\u0441\u0442\u0438 10 \u0438\u043B\u0438 13 \u0446\u0438\u0444\u0440" : "\u041D\u0435\u0442 \u043F\u043E\u0434\u0445\u043E\u0434\u044F\u0449\u0438\u0445 \u0440\u0435\u0437\u0443\u043B\u044C\u0442\u0430\u0442\u043E\u0432", [n])
      , p = g.useMemo( () => e && ka(e) ? [e] : n.length !== 10 && n.length !== 13 ? [] : i || [], [i, e, n.length])
      , u = g.useCallback( (b, y, w) => {
        const E = b == null ? void 0 : b.target;
        if (E && w === "input") {
            const A = E.value;
            (/^\d+$/.test(A) || A === "") && o(A);
        }
    }
    , [])
      , f = g.useCallback( (b, y) => {
        if (ka(y) && y) {
            const w = `${y == null ? void 0 : y.short_name_native} / \u0418\u041D\u041D: ${y == null ? void 0 : y.inn}`;
            o(w),
            t == null ? void 0 : t(y);
        }
    }
    , [t])
      , m = g.useCallback( () => {
        o(""),
        t == null ? void 0 : t(void 0);
    }
    , [t]);
    return r.jsxs(K, {
        mt: 2,
        children: [r.jsx(hu, {}), r.jsx(Ra, {
            initialValues: a,
            enableReinitialize: true,
            onSubmit: (b) => c({
                ...b,
                reason: "legal_entity_added"
            }),
            validationSchema: bu,
            children: ({handleSubmit: b}) => r.jsx(Da, {
                onSubmit: b,
                name: _n,
                id: _n,
                children: r.jsx(id, {
                    label: "\u0418\u041D\u041D \u0438\u043B\u0438 \u041E\u0413\u0420\u041D",
                    type: "text",
                    options: p,
                    value: n,
                    onInputChange: u,
                    fieldName: "legal_entity",
                    onSelect: f,
                    onClear: m,
                    noOptionsText: d,
                    isLoading: l,
                    filterOptions: (y) => y,
                    getOptionKey: (y) => y.inn,
                    renderOption: (y, w) => {
                        const {key: E, ...A} = y;
                        return r.jsx(K, {
                            component: "li",
                            ...A,
                            children: r.jsx(xu, {
                                inn: w.inn,
                                regNumber: w.ogrn,
                                shortName: w.short_name_native,
                                longName: w.long_name_native
                            })
                        }, E);
                    }
                    ,
                    getOptionLabel: (y) => y.short_name_native
                })
            })
        })]
    });
}
  , wu = ({showModal: e, setShowModal: t}) => {
    const [a,n] = g.useState()
      , [o,i] = g.useState(false)
      , {initialValues: l, setValues: s} = Ve()
      , {applicationId: c} = Fe({
        from: Ie
    })
      , {isPending: d} = zt(c)
      , p = g.useCallback(async () => {
        n(void 0),
        s(l),
        t(false);
    }
    , [l, t, s])
      , u = g.useCallback(async () => {
        i(true);
    }
    , []);
    return r.jsx(on, {
        title: "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u044E \u0432 \u043F\u0440\u043E\u0435\u043A\u0442",
        content: r.jsx(yu, {
            selectedOption: a,
            setSelectedOption: n
        }),
        showModal: e,
        acceptButtonLabel: "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C",
        refuseButtonLabel: "\u041E\u0442\u043C\u0435\u043D\u0438\u0442\u044C",
        onRefuse: p,
        onAccept: u,
        acceptButtonColor: "success",
        refuseButtonColor: "error",
        isLoading: d,
        acceptButtonFormName: _n,
        acceptButtonType: "submit",
        acceptButtonDisabled: !a,
        refuseButtonDisabled: o
    });
}
  , Rn = "no_company_reason_form"
  , Cu = De({
    no_legal_entity_reason: De().shape({
        display_name: be(),
        value: be()
    })
})
  , Su = ({setSelectedOption: e}) => {
    const {initialValues: t} = Ve()
      , {data: a} = Hn()
      , {applicationId: n} = Fe({
        from: Ie
    })
      , {mutate: o} = zt(n)
      , i = g.useMemo( () => z.filter(a == null ? void 0 : a.no_legal_entity_reason, (l) => l.value !== sa.NOT_SET), [a == null ? void 0 : a.no_legal_entity_reason]);
    return r.jsxs(K, {
        mt: 2,
        children: [r.jsxs(K, {
            mb: 2,
            children: [r.jsxs(F, {
                children: ["\u0418\u0437 \u043F\u0440\u0435\u0434\u043B\u043E\u0436\u0435\u043D\u043D\u044B\u0445 \u0432\u0430\u0440\u0438\u0430\u043D\u0442\u043E\u0432 \u0432\u044B\u0431\u0435\u0440\u0438\u0442\u0435 \u0442\u043E\u0442, \u043A\u043E\u0442\u043E\u0440\u044B\u0439 \u0441\u043E\u043E\u0442\u0432\u0435\u0442\u0441\u0442\u0432\u0443\u0435\u0442 \u0432\u0430\u0448\u0435\u0439 \u0441\u0438\u0442\u0443\u0430\u0446\u0438\u0438. \u0414\u043B\u044F \u043D\u0430\u0441 \u043E\u0442\u0441\u0443\u0442\u0441\u0442\u0432\u0438\u0435 \u044E\u0440.\u043B\u0438\u0446\u0430 \u0434\u043B\u044F \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u0439 \u044F\u0432\u043B\u044F\u0435\u0442\u0441\u044F \u0441\u0435\u0440\u044C\u0435\u0437\u043D\u044B\u043C \u043E\u0433\u0440\u0430\u043D\u0438\u0447\u0435\u043D\u0438\u0435\u043C \u0434\u043B\u044F \u0434\u0430\u043B\u044C\u043D\u0435\u0439\u0448\u0435\u0439 \u0440\u0430\u0431\u043E\u0442\u044B \u0441 \u0432\u0430\u043C\u0438. \u0411\u0435\u0437 \u0440\u0430\u0437\u0440\u0435\u0448\u0435\u043D\u0438\u044F \u044D\u0442\u043E\u0439 \u0441\u0438\u0442\u0443\u0430\u0446\u0438\u0438 \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u0438", r.jsx(F, {
                    variant: "strong",
                    mx: 0.5,
                    color: "error",
                    children: "\u043D\u0435\u0432\u043E\u0437\u043C\u043E\u0436\u043D\u044B."
                })]
            }), r.jsx(F, {
                children: "\u0412 \u0434\u0430\u043B\u044C\u043D\u0435\u0439\u0448\u0435\u043C \u0443 \u0432\u0430\u0441 \u0431\u0443\u0434\u0435\u0442 \u0432\u043E\u0437\u043C\u043E\u0436\u043D\u043E\u0441\u0442\u044C \u0434\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u044E\u0440.\u043B\u0438\u0446\u043E \u043A \u0437\u0430\u044F\u0432\u043A\u0435. \u041F\u0440\u043E\u043A\u043E\u043D\u0441\u0443\u043B\u044C\u0442\u0438\u0440\u0443\u0439\u0442\u0435\u0441\u044C \u0441 \u043D\u0430\u043C\u0438, \u043A\u0430\u043A \u043B\u0443\u0447\u0448\u0435 \u0434\u0435\u0439\u0441\u0442\u0432\u043E\u0432\u0430\u0442\u044C, \u0447\u0442\u043E\u0431\u044B \u043F\u043E\u0442\u043E\u043C \u043D\u0435 \u043F\u0440\u0438\u0448\u043B\u043E\u0441\u044C \u043F\u0435\u0440\u0435\u0434\u0435\u043B\u044B\u0432\u0430\u0442\u044C."
            })]
        }), r.jsx(Ra, {
            initialValues: t,
            enableReinitialize: true,
            onSubmit: (l) => o({
                ...l,
                reason: "no_legal_entity_reason_added"
            }),
            validationSchema: Cu,
            children: ({handleSubmit: l}) => r.jsx(Da, {
                onSubmit: l,
                name: Rn,
                id: Rn,
                children: r.jsx(Ri, {
                    fieldName: "no_legal_entity_reason",
                    label: "\u041F\u043E\u0447\u0435\u043C\u0443 \u043D\u0435\u0442 \u0432\u043E\u0437\u043C\u043E\u0436\u043D\u043E\u0441\u0442\u0438 \u0443\u043A\u0430\u0437\u0430\u0442\u044C \u044E\u0440.\u043B\u0438\u0446\u043E?",
                    choices: i,
                    sx: {
                        width: "100%"
                    },
                    onSelect: e
                })
            })
        })]
    });
}
  , Eu = ({showModal: e, setShowModal: t}) => {
    const {initialValues: a, setValues: n} = Ve()
      , [o,i] = g.useState(false)
      , [l,s] = g.useState()
      , {applicationId: c} = Fe({
        from: Ie
    })
      , {isPending: d} = zt(c)
      , p = g.useCallback(async () => {
        i(true);
    }
    , [])
      , u = g.useCallback(async () => {
        s(void 0),
        n(a),
        t(false);
    }
    , [a, t, n]);
    return r.jsx(on, {
        title: "\u041F\u043E\u0447\u0435\u043C\u0443 \u043D\u0435\u0442 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438?",
        content: r.jsx(Su, {
            setSelectedOption: s
        }),
        showModal: e,
        acceptButtonLabel: "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C",
        refuseButtonLabel: "\u041E\u0442\u043C\u0435\u043D\u0438\u0442\u044C",
        onRefuse: u,
        onAccept: p,
        acceptButtonColor: "success",
        refuseButtonColor: "error",
        isLoading: d,
        acceptButtonFormName: Rn,
        acceptButtonType: "submit",
        acceptButtonDisabled: !l,
        refuseButtonDisabled: o
    });
}
  , Au = ({isEditable: e, touched: t, errors: a, showNewCompanyModal: n, setShowNewCompanyModal: o, showNoCompanyReasonModal: i, setShowNoCompanyReasonModal: l, onClickNewCompanyButton: s, onClickNoCompanyReasonButton: c}) => r.jsxs(K, {
    sx: {
        mt: 1
    },
    children: [r.jsx(bn, {
        dataTest: "new_company_button",
        onClick: s,
        color: t.legal_entity && a.legal_entity ? "error" : "primary",
        variant: "outlined",
        disabled: !e,
        size: "small",
        children: "+ \u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u0438\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044E \u043E \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438"
    }), r.jsx(bn, {
        dataTest: "no_company_reason_button",
        onClick: c,
        color: "primary",
        variant: "text",
        sx: {
            ml: 2
        },
        disabled: !e,
        size: "small",
        children: "\u041D\u0435\u0442 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438"
    }), r.jsx(wu, {
        showModal: n,
        setShowModal: o
    }), r.jsx(Eu, {
        showModal: i,
        setShowModal: l
    })]
})
  , ku = () => {
    const [e,t] = g.useState(false)
      , [a,n] = g.useState(false)
      , {isEditable: o} = g.useContext(Lt)
      , {errors: i, touched: l} = Ve()
      , s = g.useCallback( () => {
        t(true);
    }
    , [t])
      , c = g.useCallback( () => {
        n(true);
    }
    , [n]);
    return r.jsx(Au, {
        isEditable: o,
        errors: i,
        touched: l,
        showNewCompanyModal: e,
        setShowNewCompanyModal: t,
        showNoCompanyReasonModal: a,
        setShowNoCompanyReasonModal: n,
        onClickNewCompanyButton: s,
        onClickNoCompanyReasonButton: c
    });
}
  , Wo = () => r.jsxs(K, {
    children: [r.jsxs(F, {
        fontSize: 14,
        children: ["\u041A\u043E\u043C\u043F\u0430\u043D\u0438\u044F, \u0440\u0435\u0430\u043B\u0438\u0437\u0443\u044E\u0449\u0430\u044F \u043F\u0440\u043E\u0435\u043A\u0442, \u0434\u043E\u043B\u0436\u043D\u0430 \u0431\u044B\u0442\u044C", r.jsx(F, {
            variant: "strong",
            mx: 0.5,
            color: "primary",
            children: "\u041E\u0431\u0449\u0435\u0441\u0442\u0432\u043E\u043C \u0441 \u043E\u0433\u0440\u0430\u043D\u0438\u0447\u0435\u043D\u043D\u043E\u0439 \u043E\u0442\u0432\u0435\u0442\u0441\u0442\u0432\u0435\u043D\u043D\u043E\u0441\u0442\u044C\u044E (\u041E\u041E\u041E)"
        }), "\u0438\u043B\u0438", r.jsx(F, {
            variant: "strong",
            ml: 0.5,
            color: "primary",
            children: "\u0410\u043A\u0446\u0438\u043E\u043D\u0435\u0440\u043D\u044B\u043C \u043E\u0431\u0449\u0435\u0441\u0442\u0432\u043E\u043C (\u0410\u041E)"
        }), "."]
    }), r.jsxs(F, {
        sx: {
            mt: 1
        },
        fontSize: 14,
        children: ["\u0415\u0441\u043B\u0438 \u043F\u0440\u043E\u0435\u043A\u0442", r.jsx(F, {
            variant: "strong",
            ml: 0.5,
            color: "primary",
            children: "\u043D\u0435 \u0438\u043C\u0435\u0435\u0442 \u0441\u043E\u0431\u0441\u0442\u0432\u0435\u043D\u043D\u043E\u0433\u043E \u044E\u0440.\u043B\u0438\u0446\u0430 \u0432 \u0420\u043E\u0441\u0441\u0438\u0438 \u0438\u043B\u0438 \u0440\u0435\u0430\u043B\u0438\u0437\u0443\u0435\u0442\u0441\u044F \u0432 \u0434\u0440\u0443\u0433\u0438\u0445 \u0444\u043E\u0440\u043C\u0430\u0445 (\u0410\u041D\u041E, \u0418\u041F \u0438\u043B\u0438 \u0447\u0442\u043E-\u0442\u043E \u0434\u0440\u0443\u0433\u043E\u0435)"
        }), ", \u0442\u043E \u043D\u0430\u0436\u0438\u043C\u0430\u0439\u0442\u0435 \u043D\u0430 \u043A\u043D\u043E\u043F\u043A\u0443", r.jsx(F, {
            ml: 0.5,
            color: "primary",
            component: "span",
            children: "\u041D\u0415\u0422 \u041A\u041E\u041C\u041F\u0410\u041D\u0418\u0418"
        }), ". \u0412 \u044D\u0442\u043E\u043C \u0441\u043B\u0443\u0447\u0430\u0435 \u0432\u0430\u043C \u043F\u0440\u0435\u0434\u043B\u043E\u0436\u0430\u0442 \u0432\u0432\u0435\u0441\u0442\u0438 \u0434\u0430\u043D\u043D\u044B\u0435 \u043E \u0431\u0435\u043D\u0435\u0444\u0438\u0446\u0438\u0430\u0440\u0430\u0445 \u0432\u0440\u0443\u0447\u043D\u0443\u044E."]
    })]
})
  , Tu = ({isNoLegalState: e, isWithCompany: t, isEmptyState: a}) => r.jsxs(K, {
    component: "section",
    children: [r.jsx(oa, {
        title: "\u041A\u043E\u043C\u043F\u0430\u043D\u0438\u044F \u0438 \u0431\u0435\u043D\u0435\u0444\u0438\u0446\u0438\u0430\u0440\u044B",
        supportTitle: "\u041A\u043E\u043C\u043F\u0430\u043D\u0438\u044F",
        supportContent: r.jsx(Wo, {})
    }), r.jsx(Wo, {}), a && r.jsx(ku, {}), t && r.jsx(fu, {}), e && r.jsx(gu, {})]
})
  , Pu = (e="") => {
    const {data: t} = Nt(e)
      , a = g.useMemo( () => t ? t.no_legal_entity_reason !== sa.NOT_SET : true, [t])
      , n = g.useMemo( () => t ? ka(t.legal_entity) : false, [t])
      , o = g.useMemo( () => !a && !n, [a, n]);
    return {
        isNoLegalState: a,
        isWithCompany: n,
        isEmptyState: o
    };
}
  , Ou = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {isNoLegalState: t, isWithCompany: a, isEmptyState: n} = Pu(e);
    return r.jsx(Tu, {
        isNoLegalState: t,
        isWithCompany: a,
        isEmptyState: n
    });
}
  , Ui = "application_name"
  , ju = ({isEditable: e, onEditSave: t}) => r.jsxs(K, {
    component: "section",
    children: [r.jsx(oa, {
        title: "\u041A\u043E\u043C\u043C\u0435\u0440\u0447\u0435\u0441\u043A\u043E\u0435 \u043D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportTitle: "\u041D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        supportContent: "\u041A\u0430\u043A \u0432\u044B \u043D\u0430\u0437\u044B\u0432\u0430\u0435\u0442\u0435 \u0432\u0430\u0448 \u043F\u0440\u043E\u0435\u043A\u0442 \u0434\u043B\u044F \u043A\u043B\u0438\u0435\u043D\u0442\u043E\u0432? \u0417\u0434\u0435\u0441\u044C \u043C\u043E\u0436\u043D\u043E \u0443\u043A\u0430\u0437\u0430\u0442\u044C \u043A\u043E\u043C\u043C\u0435\u0440\u0447\u0435\u0441\u043A\u043E\u0435 \u043D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u043E\u0441\u043D\u043E\u0432\u043D\u043E\u0433\u043E \u043F\u0440\u043E\u0434\u0443\u043A\u0442\u0430, \u0441\u0435\u0440\u0432\u0438\u0441\u0430, \u0443\u0441\u0443\u043B\u0433\u0438. \u0422\u0430\u043A\u0436\u0435 \u043F\u043E\u0434\u043E\u0439\u0434\u0435\u0442 \u0431\u0440\u0435\u043D\u0434\u043E\u0432\u043E\u0435 \u043D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u043A\u043E\u043C\u043F\u0430\u043D\u0438\u0438."
    }), r.jsx(Mp, {
        placeholder: "\u041F\u0440\u0438\u043C\u0435\u0440: \u0421\u0443\u043F\u0435\u0440\u0410\u043F\u043F",
        fieldName: Ui,
        onEditSave: t,
        isEditable: e
    })]
})
  , Iu = () => {
    const [e,t] = Ft(Ui)
      , {values: a} = Ve()
      , {isEditable: n} = g.useContext(Lt)
      , {applicationId: o} = Fe({
        from: Ie
    })
      , {mutate: i} = zt(o)
      , l = g.useCallback(async () => {
        e.value.trim() !== t.initialValue && i({
            ...a,
            reason: "name_changed"
        });
    }
    , [e.value, t.initialValue, i, a]);
    return r.jsx(ju, {
        isEditable: n,
        onEditSave: l
    });
}
  , _u = ({id: e="", applicationId: t="", isDeletable: a}) => {
    const {mutate: n, isPending: o, isSuccess: i} = Ar(t, e)
      , l = g.useCallback( () => {
        n();
    }
    , [n]);
    return r.jsx(Dp, {
        id: e,
        isDeletable: a,
        isDisabled: o || i,
        onDelete: l
    }, e);
}
  , Dn = "new_url_form"
  , Ru = De().shape({
    url: Ni,
    type: zi
})
  , Du = () => {
    const {data: e} = eo();
    if (e) {
        const t = z.find(e.url_type, (n) => n.value === Kt.WEB)
          , a = z.find(e.url_type, (n) => n.value === Kt.WEB);
        return {
            url: "",
            type: {
                value: (t == null ? void 0 : t.value) || Kt.WEB,
                display_name: (a == null ? void 0 : a.display_name) || Kt.WEB
            },
            application_id: ""
        };
    }
    return {
        url: "",
        type: {
            value: Kt.WEB,
            display_name: Kt.WEB
        },
        application_id: ""
    };
}
  , Fu = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {submit: t, isPending: a} = ni()
      , {data: n} = eo()
      , o = Du();
    return r.jsx(Ra, {
        initialValues: o,
        enableReinitialize: true,
        onSubmit: (i) => t({
            ...i,
            application_id: e
        }),
        validationSchema: Ru,
        children: ({handleSubmit: i}) => r.jsxs(Da, {
            onSubmit: i,
            name: Dn,
            id: Dn,
            sx: {
                mt: 2
            },
            children: [r.jsx(td, {
                label: "\u0421\u0430\u0439\u0442 \u043F\u0440\u043E\u0435\u043A\u0442\u0430 \u0438\u043B\u0438 \u043F\u043E\u0441\u0430\u0434\u043E\u0447\u043D\u0430\u044F \u0441\u0442\u0440\u0430\u043D\u0438\u0446\u0430",
                fieldName: "url",
                placeholder: "www.example.com",
                isSubmitting: a
            }), r.jsx(Ri, {
                label: "\u0427\u0442\u043E \u043F\u043E \u0441\u0441\u044B\u043B\u043A\u0435?",
                fieldName: "type",
                choices: (n == null ? void 0 : n.url_type) || [],
                sx: {
                    width: "100%",
                    mt: 2
                },
                isSubmitting: a
            })]
        })
    });
}
  , Lu = ({showModal: e, setShowModal: t}) => {
    const {isPending: a, isSuccess: n} = ni()
      , o = g.useCallback(async () => {
        t(false);
    }
    , [t]);
    return g.useEffect( () => (n && t(false),
    () => {
        t(false);
    }
    ), [n, t]),
    r.jsx(on, {
        title: "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u0441\u0441\u044B\u043B\u043A\u0443 \u043D\u0430 \u0441\u0430\u0439\u0442 \u043F\u0440\u043E\u0435\u043A\u0442\u0430",
        content: r.jsx(Fu, {}),
        showModal: e,
        acceptButtonLabel: "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C",
        refuseButtonLabel: "\u041E\u0442\u043C\u0435\u043D\u0438\u0442\u044C",
        onRefuse: o,
        acceptButtonColor: "success",
        refuseButtonColor: "error",
        isLoading: a,
        acceptButtonFormName: Dn,
        acceptButtonType: "submit"
    });
}
  , Mu = ({applicationUrls: e, applicationId: t, isEditable: a, onAddUrl: n, showModal: o, setShowModal: i}) => r.jsxs(K, {
    component: "section",
    sx: {
        mt: 3
    },
    children: [r.jsx(ca, {
        container: true,
        spacing: 2,
        children: e == null ? void 0 : e.map( (l) => r.jsx(ca, {
            size: {
                xs: 12,
                sm: 6,
                md: 4,
                lg: 3
            },
            children: r.jsx(_u, {
                applicationId: t,
                id: l.id,
                isDeletable: a
            }, l.id)
        }, l.id))
    }), a && r.jsx($n, {
        onClick: n,
        children: "+ \u0434\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u0441\u0441\u044B\u043B\u043A\u0443"
    }), r.jsx(Lu, {
        showModal: o,
        setShowModal: i
    })]
})
  , $u = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = kr(e)
      , [a,n] = g.useState(false)
      , {isEditable: o} = g.useContext(Lt)
      , i = g.useCallback( () => {
        n(true);
    }
    , [n]);
    return r.jsx(Mu, {
        applicationId: e,
        applicationUrls: t,
        isEditable: o,
        showModal: a,
        setShowModal: n,
        onAddUrl: i
    });
}
  , Bu = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = Nt(e)
      , {submit: a} = oi(e)
      , n = Fd()
      , o = g.useMemo( () => Dd(t), [t]);
    return r.jsx(Ra, {
        initialValues: n,
        enableReinitialize: true,
        onSubmit: (i) => a(i),
        validationSchema: o,
        children: ({handleSubmit: i}) => r.jsxs(Da, {
            onSubmit: i,
            name: jn,
            id: jn,
            children: [r.jsx(Iu, {}), r.jsx(Rd, {}), r.jsx($u, {}), r.jsx(Gd, {}), r.jsx(qd, {}), r.jsx(Ou, {}), r.jsx(Id, {})]
        })
    });
}
  , Nu = () => {
    const {applicationId: e} = Fe({
        from: Ie
    })
      , {data: t} = Nt(e);
    return r.jsxs(K, {
        display: "flex",
        alignItems: "center",
        children: [r.jsxs(F, {
            display: "flex",
            variant: "subtitle2",
            mt: 1,
            children: ["ID \u0437\u0430\u044F\u0432\u043A\u0438:", r.jsx(F, {
                variant: "strong",
                ml: 1,
                children: t == null ? void 0 : t.short_id
            })]
        }), (t == null ? void 0 : t.state) === Xa.NOT_SENT && r.jsx(F, {
            mt: 1,
            ml: 2,
            color: "error",
            variant: "strong",
            children: "\u0417\u0430\u044F\u0432\u043A\u0430 \u043D\u0435 \u043F\u043E\u0434\u0430\u043D\u0430"
        })]
    });
}
  , zu = ({pageTitle: e, isSubmitting: t, isEditable: a, showModal: n, setShowModal: o, applicationState: i}) => r.jsxs(K, {
    display: "flex",
    flexDirection: "column",
    sx: {
        mb: 5
    },
    children: [r.jsx(Fp, {
        children: e
    }), r.jsx(Nu, {}), r.jsx(Bu, {}), i === Xa.NOT_SENT && r.jsx(Pr, {
        form: jn,
        label: "\u041F\u043E\u0434\u0430\u0442\u044C \u0437\u0430\u044F\u0432\u043A\u0443",
        color: "success",
        disabled: t || !a,
        isSubmitting: t,
        sx: {
            ml: "auto",
            width: 250
        }
    }), r.jsx(Oi, {
        title: "\u041F\u043E\u0437\u0434\u0440\u0430\u0432\u043B\u044F\u0435\u043C \u0441 \u043F\u043E\u0434\u0430\u0447\u0435\u0439 \u0437\u0430\u044F\u0432\u043A\u0438",
        content: "\u041F\u043E\u0441\u043B\u0435 \u043F\u043E\u0434\u0430\u0447\u0438 \u0437\u0430\u044F\u0432\u043A\u0438 \u0440\u0430\u0431\u043E\u0442\u0430 \u043D\u0430\u0434 \u043D\u0435\u0439 \u043D\u0435 \u043F\u0440\u0435\u043A\u0440\u0430\u0449\u0430\u0435\u0442\u0441\u044F. \u041C\u044B \u0431\u0443\u0434\u0435\u043C \u0434\u043E\u0437\u0430\u043F\u0440\u0430\u0448\u0438\u0432\u0430\u0442\u044C \u043D\u0435\u043E\u0431\u0445\u043E\u0434\u0438\u043C\u0443\u044E \u0434\u043B\u044F \u043F\u0440\u0438\u043D\u044F\u0442\u0438\u044F \u0440\u0435\u0448\u0435\u043D\u0438\u044F \u0438\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044E \u0447\u0435\u0440\u0435\u0437 \u0438\u043D\u0442\u0435\u0440\u0444\u0435\u0439\u0441 \u0437\u0430\u044F\u0432\u043A\u0438. \u041F\u0440\u043E\u0432\u0435\u0440\u044F\u0439\u0442\u0435 \u043F\u043E\u0447\u0442\u0443, \u0442\u0430\u043C \u0432\u0430\u0441 \u0431\u0443\u0434\u0443\u0442 \u0438\u043D\u0444\u043E\u0440\u043C\u0438\u0440\u043E\u0432\u0430\u0442\u044C \u043E\u0431 \u0438\u0437\u043C\u0435\u043D\u0435\u043D\u0438\u044F\u0445.",
        showModal: n,
        setShowModal: o
    })]
})
  , Vu = () => {
    const [e,t] = g.useState(false)
      , {applicationId: a} = Fe({
        from: Ie
    })
      , {data: n} = Nt(a)
      , {isSuccess: o, isPending: i} = oi(a);
    Fr(`${n == null ? void 0 : n.name}` || "\u0427\u0435\u0440\u043D\u043E\u0432\u0438\u043A \u0437\u0430\u044F\u0432\u043A\u0438");
    const l = g.useMemo( () => (n == null ? void 0 : n.state) === Xa.NOT_SENT ? "\u0427\u0435\u0440\u043D\u043E\u0432\u0438\u043A \u0437\u0430\u044F\u0432\u043A\u0438" : (n == null ? void 0 : n.state) === Xa.SUBMITTED ? "\u0417\u0430\u044F\u0432\u043A\u0430 \u043D\u0430 \u0438\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u0438" : "", [n]);
    return g.useEffect( () => {
        o && t(true);
    }
    , [o]),
    r.jsx(fd, {
        children: r.jsx(zu, {
            pageTitle: l,
            isSubmitting: i || o,
            isEditable: n == null ? void 0 : n.is_editable,
            showModal: e,
            setShowModal: t,
            applicationState: n == null ? void 0 : n.state
        })
    });
}
  , im = hr(Ie)({
    component: Vu
});
export {im as Route};
