document.addEventListener("DOMContentLoaded", function () {
    const exchangeRates = {
        usd: 1,
        eur: 0.93,
        rub: 90
    };

    const currencySymbols = {
        usd: "$",
        eur: "€",
        rub: "₽"
    };

    const planExtras = {
        Startup: "-",
        Basic: "Tehnical support",
        Standard: "Technical support, API access, Telegram alerts",
        Business: "API, Telegram alerts, Priority support",
        Pro: "LDAP/AD, Monitoring, SLA, API, Telegram",
        Enterprise: "WAF, Zero Trust, SLA, Dedicated manager, Custom integrations"
    };
    
    

    let currency = "usd"; // can be changed dynamically via selector

    function determinePlan(users) {
        if (users <= 5) return "Startup";
        if (users <= 10) return "Basic";
        if (users <= 30) return "Standard";
        if (users <= 60) return "Business";
        if (users <= 120) return "Pro";
        return "Enterprise";
    }

    const perUserPrices = {
        "Startup": 0,
        "Basic": 3,
        "Standard": 3,
        "Business": 3,
        "Pro": 3,
        "Enterprise": 3
    };

    const userLimits = {
        "Startup": 5,
        "Basic": 10,
        "Standard": 30,
        "Business": 60,
        "Pro": 120,
        "Enterprise": Infinity
    };

    const serverLimits = {
        "Startup": 1,
        "Basic": 3,
        "Standard": 10,
        "Business": 25,
        "Pro": 50,
        "Enterprise": Infinity
    };

    const groupLimits = {
        "Startup": 1,
        "Basic": 3,
        "Standard": 5,
        "Business": 15,
        "Pro": 30,
        "Enterprise": Infinity
    };

    const planBenefits = {
        "Basic": "Email support",
        "Standard": "API access, Telegram alerts",
        "Business": "Priority support",
        "Pro": "LDAP, monitoring",
        "Enterprise": "WAF, Zero Trust, SLA, Account Manager"
    };

    function calculateExtraCost(actual, limit, rateSmall, rateLarge) {
        const extra = actual - limit;
        if (extra <= 0) return 0;
        if (extra <= 3) return extra * rateSmall;
        return (3 * rateSmall) + ((extra - 3) * rateLarge);
    }

    function formatPrice(amount) {
        const converted = amount * exchangeRates[currency];
        return `${converted.toFixed(2)} ${currencySymbols[currency]}`;
    }
    
    // 👇 Вставь эту функцию прямо после formatPrice
    function updateCurrencyLabels() {
        const rate = exchangeRates[currency];
        const symbol = currencySymbols[currency];
    
        // Add-ons
        document.querySelectorAll(".addons-toggle label").forEach(label => {
            const fixed = label.getAttribute("data-price-usd");
            const perUser = label.getAttribute("data-price-usd-per-user");
    
            if (fixed) {
                const value = (parseFloat(fixed) * rate).toFixed(2);
                label.textContent = `${label.htmlFor.toUpperCase()} (+${value} ${symbol}/month)`;
            } else if (perUser) {
                const value = (parseFloat(perUser) * rate).toFixed(2);
                label.textContent = `${label.htmlFor.toUpperCase()} (+${value} ${symbol}/month per user)`;
            }
        });
    
        // Plan cards
        document.querySelectorAll(".plan-price[data-price-usd]").forEach(span => {
            const usdPrice = parseFloat(span.getAttribute("data-price-usd"));
            const converted = (usdPrice * rate).toFixed(2);
            span.textContent = `~${converted}${symbol}/user`;
        });
    }
    

    function updateSliderBackground(slider, maxLimit) {
        if (!slider) return;
    
        const value = parseInt(slider.value);
        const max = parseInt(slider.max);
    
        const valuePercent = (value / max) * 100;
        const limitPercent = (maxLimit / max) * 100;
    
        let gradient;
    
        if (value <= maxLimit) {
            // Всё в пределах лимита - зелёный
            gradient = `linear-gradient(to right, #4CAF50 0%, #4CAF50 ${valuePercent}%, #ddd ${valuePercent}%, #ddd 100%)`;
        } else {
            // Превышен лимит: зелёный до лимита, красный до значения, дальше серый
            gradient = `linear-gradient(
                to right, 
                #4CAF50 0%, 
                #4CAF50 ${limitPercent}%, 
                #ff4d4d ${limitPercent}%, 
                #ff4d4d ${valuePercent}%, 
                #ddd ${valuePercent}%, 
                #ddd 100%
            )`;
        }
    
        slider.style.background = gradient;
    }

    function animateSliderToQuota(slider, targetValue) {
        const current = parseInt(slider.value);
        const step = current < targetValue ? 1 : -1;
    
        const interval = setInterval(() => {
            const value = parseInt(slider.value);
            if (value === targetValue) {
                clearInterval(interval);
            } else {
                slider.value = value + step;
                slider.dispatchEvent(new Event("input")); // чтобы обновлялась цена и фон
            }
        }, 10); // скорость анимации
    }

    document.querySelectorAll("input[name='plan']").forEach(input => {
        input.addEventListener("change", function () {
            const selected = this.value;
    
            // 1. Очистить старые выделения
            document.querySelectorAll(".toggle_price label").forEach(label => {
                label.classList.remove("selected-plan");
            });
    
            // 2. Добавить .selected-plan к текущей
            const label = document.querySelector(`label[for="${this.id}"]`);
            if (label) {
                label.classList.add("selected-plan");
            }
    
            // 3. Найти квоты по выбранному тарифу
            const targetUsers = userLimits[selected];
            const targetServers = serverLimits[selected];
            const targetGroups = groupLimits[selected];
    
            // 4. Анимировать ползунки до лимитов
            animateSliderToQuota(document.querySelector("input[name='users']"), targetUsers === Infinity ? 121 : targetUsers);
            animateSliderToQuota(document.querySelector("input[name='servers']"), targetServers === Infinity ? 100 : targetServers);
            animateSliderToQuota(document.querySelector("input[name='groups']"), targetGroups === Infinity ? 50 : targetGroups);
        });
    });
    
    
    
    

    function calculatePrice() {
        let users = parseInt(document.querySelector("input[name='users']")?.value) || 1;
        let servers = parseInt(document.querySelector("input[name='servers']")?.value) || 1;
        let groups = parseInt(document.querySelector("input[name='groups']")?.value) || 1;

        let ldap = document.querySelector("#ldap")?.checked ? 30 : 0;
        let monitoring = document.querySelector("#monitoring")?.checked ? 50 : 0;
        let waf = document.querySelector("#waf")?.checked ? 70 : 0;
        let zeroTrust = document.querySelector("#zero-trust")?.checked ? users * 6.5 : 0;

        let plan = determinePlan(users);
        let perUser = perUserPrices[plan] || 0;

        let planInput = document.querySelector(`input[value='${plan}']`);
        if (planInput) {
            planInput.checked = true;
        
            // сброс старых выделений
            document.querySelectorAll(".toggle_price label").forEach(label => {
                label.classList.remove("selected-plan");
            });
        
            // установка новой карточки
            const label = document.querySelector(`label[for="${planInput.id}"]`);
            if (label) {
                label.classList.add("selected-plan");
            }
        }
        

        let planNameElem = document.getElementById("selectedPlanName");
        if (planNameElem) planNameElem.textContent = plan;

        let userPrice = users * perUser;
        let extraServerPrice = calculateExtraCost(servers, serverLimits[plan], 5, 10);
        let extraGroupPrice = calculateExtraCost(groups, groupLimits[plan], 3, 7);
        let addonsTotalPrice = ldap + monitoring + waf + zeroTrust;

        let total = userPrice + extraServerPrice + extraGroupPrice + addonsTotalPrice;

        // Suggested upgrade
        let suggestedPlan = null;
        let savings = 0;
        let plans = ["Basic", "Standard", "Business", "Pro", "Enterprise"];
        for (let i = 0; i < plans.length; i++) {
            const candidate = plans[i];
            const limit = userLimits[candidate];
            if (users <= limit && candidate !== plan) {
                let newUserPrice = users * perUserPrices[candidate];
                let oldTotal = userPrice + extraServerPrice + extraGroupPrice;
                let newTotal = newUserPrice;
                if (newTotal < oldTotal) {
                    suggestedPlan = candidate;
                    savings = oldTotal - newTotal;
                }
                break;
            }
            updateCurrencyLabels();

        }

        document.getElementById("amount").innerHTML = formatPrice(total) + " / month";
        document.getElementById("userCount").textContent = users >= 999 ? "∞" : users;
        document.getElementById("serverCount").textContent = servers;
        document.getElementById("groupCount").textContent = groups;

        document.getElementById("quotaUserCount").textContent = userLimits[plan] === Infinity ? "∞" : userLimits[plan];
        document.getElementById("quotaServerCount").textContent = serverLimits[plan] === Infinity ? "∞" : serverLimits[plan];
        document.getElementById("quotaGroupCount").textContent = groupLimits[plan] === Infinity ? "∞" : groupLimits[plan];

        document.getElementById("extra-costs-servers").textContent = formatPrice(extraServerPrice);
        document.getElementById("extra-costs-groups").textContent = formatPrice(extraGroupPrice);

        const benefitElem = document.getElementById("plan-benefits");
        const suggestionElem = document.getElementById("plan-suggestion");
        if (benefitElem) {
            benefitElem.textContent = planExtras[plan] || "-";
        }
        
        if (suggestionElem) {
            suggestionElem.innerHTML = suggestedPlan
                ? `We recommend switching to <b>${suggestedPlan}</b> – you save <b>${formatPrice(savings)}</b>`
                : "-";
        }

        let addonsList = document.getElementById("addons-list");
        let addonsTotalElem = document.getElementById("addons-total");

        if (addonsList && addonsTotalElem) {
            addonsList.innerHTML = "";
            addonsTotalElem.textContent = formatPrice(addonsTotalPrice);

            function addAddon(name, price) {
                if (price > 0) {
                    let li = document.createElement("li");
                    li.textContent = `${name} (+${formatPrice(price)})`;
                    addonsList.appendChild(li);
                }
            }

            addAddon("LDAP", ldap);
            addAddon("Monitoring", monitoring);
            addAddon("WAF", waf);
            addAddon("Zero Trust", zeroTrust);
        }

        updateSliderBackground(document.querySelector("input[name='users']"), userLimits[plan]);
        updateSliderBackground(document.querySelector("input[name='servers']"), serverLimits[plan]);
        updateSliderBackground(document.querySelector("input[name='groups']"), groupLimits[plan]);

        document.querySelectorAll(".addons-toggle input[type='checkbox']").forEach((checkbox) => {
            let label = document.querySelector(`label[for='${checkbox.id}']`);
            if (label) {
                label.classList.toggle("checked-addon", checkbox.checked);
            }
        });

        console.log("Plan:", plan);
        console.log("Currency:", currency);
    }

    document.querySelectorAll("input[type='range'], input[type='checkbox']").forEach(input => {
        input.addEventListener("input", calculatePrice);
        input.addEventListener("change", calculatePrice);
    });

    // слушатель переключения валют
    document.querySelectorAll(".currency-switcher input[type='radio']").forEach(input => {
        input.addEventListener("change", function () {
            currency = this.value;
            calculatePrice();
            updateCurrencyLabels();
        });
    });
    

    
        

    document.addEventListener("DOMContentLoaded", function () {
        document.querySelectorAll(".addons-toggle label").forEach(label => {
            label.addEventListener("mouseenter", function () {
                this.classList.add("tooltip-visible");
            });
            label.addEventListener("mouseleave", function () {
                this.classList.remove("tooltip-visible");
            });
        });
    });

    calculatePrice();
});
